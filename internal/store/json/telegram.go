package jsonstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/domain"
)

type TelegramQuestion struct {
	QuestionID string `json:"question_id"`
	UserHash   string `json:"user_hash"`
	BotHash    string `json:"bot_hash"`
	UpdateID   int64  `json:"update_id"`
	CreatedAt  string `json:"created_at"`
}

type TelegramQuestionResult string

type TelegramPoll struct {
	BotHash      string `json:"bot_hash"`
	LastUpdateID int64  `json:"last_update_id"`
}

func loadTelegramPoll(path string) (TelegramPoll, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return TelegramPoll{}, nil
	}
	if err != nil {
		return TelegramPoll{}, err
	}
	var state TelegramPoll
	if err := json.Unmarshal(b, &state); err != nil {
		return state, fmt.Errorf("parse %s: %w", path, err)
	}
	return state, nil
}

func (s *Store) TelegramOffset(token string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.telegramPoll.BotHash != botHash(token) {
		return 0
	}
	return s.telegramPoll.LastUpdateID + 1
}

func (s *Store) RecordTelegramUpdate(token string, updateID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := botHash(token)
	if s.telegramPoll.BotHash == hash && s.telegramPoll.LastUpdateID >= updateID {
		return nil
	}
	next := TelegramPoll{BotHash: hash, LastUpdateID: updateID}
	if err := saveJSON(config.TelegramPollPath, next); err != nil {
		return err
	}
	s.telegramPoll = next
	return nil
}

const (
	TelegramQuestionCreated    TelegramQuestionResult = "created"
	TelegramQuestionDuplicate  TelegramQuestionResult = "duplicate"
	TelegramQuestionDailyLimit TelegramQuestionResult = "daily_limit"
	TelegramQuestionPending    TelegramQuestionResult = "pending"
)

// Legacy Telegram fields are removed from the general settings file once the
// independent Telegram settings file has been written.
func loadTelegramSettings(path, legacyPath string) (domain.TelegramSettings, error) {
	var settings domain.TelegramSettings
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, &settings); err != nil {
			return settings, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		settings.DailyLimit = 10
		legacy, err := os.ReadFile(legacyPath)
		if err != nil {
			return settings, err
		}
		var old struct {
			Enabled  *bool  `json:"telegram_enabled"`
			BotToken string `json:"telegram_bot_token"`
			UserID   string `json:"telegram_user_id"`
		}
		if err := json.Unmarshal(legacy, &old); err != nil {
			return settings, err
		}
		settings.BotToken, settings.UserID = old.BotToken, old.UserID
		settings.PushEnabled = old.Enabled != nil && *old.Enabled || old.Enabled == nil && old.BotToken != "" && old.UserID != ""
		if err := saveJSON(path, settings); err != nil {
			return settings, err
		}
	} else {
		return settings, err
	}
	settings.BotToken = strings.TrimSpace(settings.BotToken)
	settings.UserID = strings.TrimSpace(settings.UserID)
	if settings.DailyLimit <= 0 {
		settings.DailyLimit = 10
	}
	legacy, err := os.ReadFile(legacyPath)
	if err == nil && (strings.Contains(string(legacy), `"telegram_enabled"`) || strings.Contains(string(legacy), `"telegram_bot_token"`) || strings.Contains(string(legacy), `"telegram_user_id"`)) {
		var general domain.Settings
		if err := json.Unmarshal(legacy, &general); err != nil {
			return settings, err
		}
		if err := saveJSON(legacyPath, general); err != nil {
			return settings, err
		}
	}
	return settings, nil
}

func loadTelegramQuestions(path string) ([]TelegramQuestion, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		items := []TelegramQuestion{}
		return items, saveJSON(path, items)
	}
	if err != nil {
		return nil, err
	}
	var items []TelegramQuestion
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if items == nil {
		items = []TelegramQuestion{}
	}
	return items, nil
}

func TelegramUserHash(userID int64) string {
	sum := sha256.Sum256([]byte("T" + strconv.FormatInt(userID, 10)))
	return hex.EncodeToString(sum[:])
}

func botHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Store) AddTelegramQuestion(userID, updateID int64, botToken, content string) (domain.Question, TelegramQuestionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userHash, currentBotHash := TelegramUserHash(userID), botHash(botToken)
	for _, item := range s.telegramQuestions {
		if item.BotHash == currentBotHash && item.UpdateID == updateID {
			for _, q := range s.questions {
				if q.ID == item.QuestionID {
					return q, TelegramQuestionDuplicate, nil
				}
			}
			// A crash after writing the mapping but before writing the question is repaired below.
			q := domain.Question{ID: item.QuestionID, Content: strings.TrimSpace(content), Status: "pending", CreatedAt: item.CreatedAt}
			if err := s.saveTelegramQuestionRecord(q); err != nil {
				return domain.Question{}, "", err
			}
			return q, TelegramQuestionCreated, nil
		}
	}
	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	count := 0
	statusByID := make(map[string]string, len(s.questions))
	for _, q := range s.questions {
		statusByID[q.ID] = q.Status
	}
	for _, item := range s.telegramQuestions {
		if strings.HasPrefix(item.CreatedAt, today) {
			count++
		}
		if !s.telegram.AllowMultiplePending && item.UserHash == userHash {
			if statusByID[item.QuestionID] == "pending" {
				return domain.Question{}, TelegramQuestionPending, nil
			}
		}
	}
	if count >= s.telegram.DailyLimit {
		return domain.Question{}, TelegramQuestionDailyLimit, nil
	}
	q := domain.Question{ID: newID(), Content: strings.TrimSpace(content), Status: "pending", CreatedAt: now.Format(time.RFC3339)}
	item := TelegramQuestion{QuestionID: q.ID, UserHash: userHash, BotHash: currentBotHash, UpdateID: updateID, CreatedAt: q.CreatedAt}
	updated := append(append([]TelegramQuestion(nil), s.telegramQuestions...), item)
	if err := saveJSON(config.TelegramQuestionsPath, updated); err != nil {
		return domain.Question{}, "", err
	}
	s.telegramQuestions = updated
	if err := s.saveTelegramQuestionRecord(q); err != nil {
		return domain.Question{}, "", err
	}
	return q, TelegramQuestionCreated, nil
}

func (s *Store) saveTelegramQuestionRecord(q domain.Question) error {
	updated := append([]domain.Question{q}, s.questions...)
	if err := saveJSON(config.QuestionsPath, updated); err != nil {
		return err
	}
	s.questions = updated
	return nil
}

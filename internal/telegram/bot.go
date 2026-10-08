package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
	jsonstore "github.com/Magnetkopf/PAB-Go/internal/store/json"
)

type BotStore interface {
	TelegramSettings() domain.TelegramSettings
	TelegramOffset(string) int64
	RecordTelegramUpdate(string, int64) error
	AddTelegramQuestion(int64, int64, string, string) (domain.Question, jsonstore.TelegramQuestionResult, error)
}

type update struct {
	ID      int64 `json:"update_id"`
	Message *struct {
		Text string `json:"text"`
		Date int64  `json:"date"`
		From *struct {
			ID    int64 `json:"id"`
			IsBot bool  `json:"is_bot"`
		} `json:"from"`
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
	} `json:"message"`
}

func (c *Client) getUpdates(ctx context.Context, token string, offset int64) ([]update, error) {
	values := url.Values{"timeout": {"25"}, "offset": {strconv.FormatInt(offset, 10)}, "allowed_updates": {`["message"]`}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL(token, "getUpdates"), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := *c.httpClient
	client.Timeout = 35 * time.Second
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Telegram polling request failed")
	}
	defer response.Body.Close()
	var result struct {
		OK          bool     `json:"ok"`
		Description string   `json:"description"`
		Result      []update `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode Telegram updates: %w", err)
	}
	if response.StatusCode != http.StatusOK || !result.OK {
		return nil, fmt.Errorf("Telegram polling API: %s", result.Description)
	}
	return result.Result, nil
}

func (c *Client) Run(ctx context.Context, store BotStore) {
	for ctx.Err() == nil {
		settings := store.TelegramSettings()
		if !settings.AskEnabled || settings.BotToken == "" {
			pause(ctx, 2*time.Second)
			continue
		}
		pollCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
		updates, err := c.getUpdates(pollCtx, settings.BotToken, store.TelegramOffset(settings.BotToken))
		cancel()
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("Telegram polling: %v", err)
				pause(ctx, 5*time.Second)
			}
			continue
		}
		for _, item := range updates {
			// A token or ask setting change takes effect before the next update.
			current := store.TelegramSettings()
			if !current.AskEnabled || current.BotToken != settings.BotToken {
				break
			}
			if err := c.handleUpdate(ctx, store, current, item); err != nil {
				log.Printf("Telegram update: %v", err)
				break
			}
			if err := store.RecordTelegramUpdate(settings.BotToken, item.ID); err != nil {
				log.Printf("save Telegram update cursor: %v", err)
				break
			}
		}
	}
}

func (c *Client) handleUpdate(ctx context.Context, store BotStore, settings domain.TelegramSettings, item update) error {
	m := item.Message
	if m == nil || m.Chat.Type != "private" || m.From == nil || m.From.IsBot || m.From.ID == 0 {
		return nil
	}
	if settings.AskEnabledAt > 0 && m.Date > 0 && m.Date < settings.AskEnabledAt {
		return nil
	}
	chatID := strconv.FormatInt(m.Chat.ID, 10)
	message := strings.TrimSpace(m.Text)
	if strings.HasPrefix(message, "/ask ") {
		message = strings.TrimSpace(strings.TrimPrefix(message, "/ask "))
	}
	var reply string
	switch {
	case message == "/start" || message == "/help":
		reply = "Send me a text question or use /ask followed by your question (5 to 1,000 characters). I will add it to the inbox. Questions are anonymous on the website."
	case message == "/ask":
		reply = "Send /ask followed by your question, or just send your question as a message."
	case strings.HasPrefix(message, "/"):
		reply = "Unknown command. Send /help for instructions."
	case message == "":
		reply = "Please send a text question. Images and other attachments are not supported yet."
	case len([]rune(message)) < 5:
		reply = "Your question is too short. Please use at least 5 characters."
	case len([]rune(message)) > 1000:
		reply = "Your question is too long. Please keep it within 1,000 characters."
	default:
		question, outcome, err := store.AddTelegramQuestion(m.From.ID, item.ID, settings.BotToken, message)
		if err != nil {
			return err
		}
		switch outcome {
		case jsonstore.TelegramQuestionCreated:
			reply = "Your question has been added to the inbox."
			if settings.PushEnabled && settings.UserID != "" {
				if err := c.SendQuestion(ctx, settings, question, ""); err != nil {
					log.Printf("notify Telegram owner: %v", err)
				}
			}
		case jsonstore.TelegramQuestionDuplicate:
			reply = "Your question was already received."
		case jsonstore.TelegramQuestionPending:
			reply = "You already have a question waiting for an answer. Please wait before asking another."
		case jsonstore.TelegramQuestionDailyLimit:
			reply = "The question limit for today has been reached. Please try again tomorrow (UTC)."
		}
	}
	// A failed acknowledgement must not cause an accepted question to be inserted twice.
	if err := c.SendChatText(ctx, settings.BotToken, chatID, reply); err != nil {
		log.Printf("send Telegram reply: %v", err)
	}
	return nil
}

func pause(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

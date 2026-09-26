package jsonstore

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	settings  domain.Settings
	questions []domain.Question
	sessions  []domain.Session
}

// New ensures the fixed data files exist. Invalid JSON is returned to main,
// which intentionally panics rather than running with possibly corrupt data.
func New() (*Store, error) {
	settings, err := loadSettings(config.SettingsPath)
	if err != nil {
		return nil, err
	}
	questions, err := loadQuestions(config.QuestionsPath)
	if err != nil {
		return nil, err
	}
	sessions, err := loadSessions(config.SessionsPath)
	if err != nil {
		return nil, err
	}
	return &Store{settings: settings, questions: questions, sessions: sessions}, nil
}

func loadSettings(path string) (domain.Settings, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		settings := defaultSettings()
		return settings, saveJSON(path, settings)
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("read %s: %w", path, err)
	}
	var settings domain.Settings
	if err := json.Unmarshal(b, &settings); err != nil {
		return domain.Settings{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return normalizeSettings(settings), nil
}

func loadQuestions(path string) ([]domain.Question, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		questions := []domain.Question{}
		return questions, saveJSON(path, questions)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var questions []domain.Question
	if err := json.Unmarshal(b, &questions); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if questions == nil {
		questions = []domain.Question{}
	}
	return questions, nil
}

func loadSessions(path string) ([]domain.Session, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		sessions := []domain.Session{}
		return sessions, saveJSON(path, sessions)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var sessions []domain.Session
	if err := json.Unmarshal(b, &sessions); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if sessions == nil {
		sessions = []domain.Session{}
	}
	return sessions, nil
}

func (s *Store) Settings() domain.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Store) UpdateSettings(next domain.Settings) (domain.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = normalizeSettings(next)
	return s.settings, saveJSON(config.SettingsPath, s.settings)
}

func (s *Store) AddQuestion(nickname, content, imageFilename string) (domain.Question, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := domain.Question{ID: newID(), Nickname: strings.TrimSpace(nickname), Content: strings.TrimSpace(content), ImageFilename: imageFilename, Status: "pending", CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	s.questions = append([]domain.Question{q}, s.questions...)
	return q, saveJSON(config.QuestionsPath, s.questions)
}

func (s *Store) Questions(status string) []domain.Question {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Question, 0, len(s.questions))
	for _, q := range s.questions {
		if status == "" || status == "all" || q.Status == status {
			out = append(out, q)
		}
	}
	return out
}

func (s *Store) Search(query string) []domain.Question {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query = strings.ToLower(strings.TrimSpace(query))
	out := []domain.Question{}
	for _, q := range s.questions {
		if q.Status == "published" && strings.Contains(strings.ToLower(q.Content+" "+q.Answer+" "+q.Nickname), query) {
			out = append(out, q)
		}
	}
	return out
}

func (s *Store) Answer(id, answer string, publish bool) (domain.Question, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.questions {
		if s.questions[i].ID != id {
			continue
		}
		q := &s.questions[i]
		q.Answer = strings.TrimSpace(answer)
		q.AnsweredAt = time.Now().UTC().Format(time.RFC3339)
		if publish {
			q.Status = "published"
		} else {
			q.Status = "answered"
		}
		return *q, saveJSON(config.QuestionsPath, s.questions)
	}
	return domain.Question{}, os.ErrNotExist
}

func (s *Store) CreateSession(tokenHash string, expiresAt time.Time) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropExpiredSessions(time.Now())
	session := domain.Session{ID: newID(), TokenHash: tokenHash, CreatedAt: time.Now().UTC(), ExpiresAt: expiresAt.UTC()}
	s.sessions = append([]domain.Session{session}, s.sessions...)
	return session, saveJSON(config.SessionsPath, s.sessions)
}

func (s *Store) SessionByTokenHash(tokenHash string) (domain.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	for _, session := range s.sessions {
		if session.ExpiresAt.After(now) && subtle.ConstantTimeCompare([]byte(session.TokenHash), []byte(tokenHash)) == 1 {
			return session, true
		}
	}
	return domain.Session{}, false
}

func (s *Store) Sessions() []domain.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	result := make([]domain.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		if session.ExpiresAt.After(now) {
			result = append(result, session)
		}
	}
	return result
}

func (s *Store) RevokeSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropExpiredSessions(time.Now())
	for i, session := range s.sessions {
		if session.ID == id {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return saveJSON(config.SessionsPath, s.sessions)
		}
	}
	return os.ErrNotExist
}

func (s *Store) RevokeAllSessions() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sessions) == 0 {
		return nil
	}
	s.sessions = []domain.Session{}
	return saveJSON(config.SessionsPath, s.sessions)
}

func (s *Store) dropExpiredSessions(now time.Time) {
	active := s.sessions[:0]
	for _, session := range s.sessions {
		if session.ExpiresAt.After(now) {
			active = append(active, session)
		}
	}
	s.sessions = active
}

func saveJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".pab-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func defaultSettings() domain.Settings {
	return domain.Settings{SiteName: "AskBox", PrimaryColor: "#70d7eb", CopyrightName: "Nekro", TopBarOpacity: 50, NavigationOpacity: 50, CardOpacity: 50, MaxUploadKB: 1024}
}

func normalizeSettings(s domain.Settings) domain.Settings {
	d := defaultSettings()
	if strings.TrimSpace(s.SiteName) == "" {
		s.SiteName = d.SiteName
	}
	if strings.TrimSpace(s.CopyrightName) == "" {
		s.CopyrightName = d.CopyrightName
	}
	if len(s.PrimaryColor) != 7 || s.PrimaryColor[0] != '#' {
		s.PrimaryColor = d.PrimaryColor
	}
	s.TopBarOpacity = clamp(s.TopBarOpacity)
	s.NavigationOpacity = clamp(s.NavigationOpacity)
	s.CardOpacity = clamp(s.CardOpacity)
	if s.MaxUploadKB <= 0 {
		s.MaxUploadKB = d.MaxUploadKB
	}
	if s.MaxUploadKB > 10240 {
		s.MaxUploadKB = 10240
	}
	return s
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func newID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

package jsonstore

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/domain"
)

func telegramTestDir(t *testing.T) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Error(err)
		}
	})
}

func TestLegacyTelegramSettingsMigrateToSeparateFile(t *testing.T) {
	telegramTestDir(t)
	if err := os.MkdirAll(config.DataDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.SettingsPath, []byte(`{"site_name":"Site","telegram_bot_token":"token","telegram_user_id":"1234"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if !s.TelegramSettings().PushEnabled || s.TelegramSettings().BotToken != "token" {
		t.Fatalf("migration = %+v", s.TelegramSettings())
	}
	b, err := os.ReadFile(config.SettingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "telegram_") {
		t.Fatalf("Telegram settings remain in general settings: %s", b)
	}
	b, err = os.ReadFile(config.TelegramSettingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "token") {
		t.Fatalf("separate settings missing token: %s", b)
	}
}

func TestExplicitlyDisabledLegacyTelegramStaysDisabled(t *testing.T) {
	telegramTestDir(t)
	if err := os.MkdirAll(config.DataDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.SettingsPath, []byte(`{"telegram_enabled":false,"telegram_bot_token":"token","telegram_user_id":"1234"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if s.TelegramSettings().PushEnabled {
		t.Fatal("legacy notifications were enabled despite explicit false")
	}
}

func TestTelegramQuestionHashLimitAndPersistence(t *testing.T) {
	telegramTestDir(t)
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	settings := domain.TelegramSettings{BotToken: "test-token", AskEnabled: true, DailyLimit: 1}
	if _, err := s.UpdateTelegramSettings(settings); err != nil {
		t.Fatal(err)
	}
	if s.TelegramSettings().AskEnabledAt == 0 {
		t.Fatal("Telegram ask enable time was not recorded")
	}
	const questionText = `<img src=x onerror=alert(1)>`
	q, outcome, err := s.AddTelegramQuestion(123456789, 7, "test-token", questionText)
	if err != nil || outcome != TelegramQuestionCreated {
		t.Fatalf("first: %s %v", outcome, err)
	}
	if q.Nickname != "" || q.Status != "pending" || q.Content != questionText {
		t.Fatalf("question = %+v", q)
	}
	_, outcome, err = s.AddTelegramQuestion(123456789, 7, "test-token", "First question")
	if err != nil || outcome != TelegramQuestionDuplicate {
		t.Fatalf("duplicate: %s %v", outcome, err)
	}
	_, outcome, err = s.AddTelegramQuestion(123456789, 8, "test-token", "Second question")
	if err != nil || outcome != TelegramQuestionPending {
		t.Fatalf("pending: %s %v", outcome, err)
	}
	_, outcome, err = s.AddTelegramQuestion(222, 9, "test-token", "Another question")
	if err != nil || outcome != TelegramQuestionDailyLimit {
		t.Fatalf("daily limit: %s %v", outcome, err)
	}
	b, err := os.ReadFile(config.TelegramQuestionsPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "123456789") || !strings.Contains(string(b), q.ID) {
		t.Fatalf("mapping contents: %s", b)
	}
	b, err = os.ReadFile(config.QuestionsPath)
	if err != nil {
		t.Fatal(err)
	}
	var questions []map[string]any
	if err := json.Unmarshal(b, &questions); err != nil {
		t.Fatal(err)
	}
	if questions[0]["content"] != questionText {
		t.Fatalf("stored content = %q", questions[0]["content"])
	}
	if _, ok := questions[0]["user_hash"]; ok {
		t.Fatal("question contains Telegram identity")
	}
	if err := os.WriteFile(config.SecretsPath, []byte(`{"session_secret":"changed"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordTelegramUpdate("test-token", 7); err != nil {
		t.Fatal(err)
	}
	reloaded, err := New()
	if err != nil {
		t.Fatal(err)
	}
	const expectedHash = "8006843456a6804e9889b5e1fe1667d7c998537e743f897d354104830a02e49b"
	if TelegramUserHash(123456789) != expectedHash {
		t.Fatalf("Telegram hash = %s, want %s", TelegramUserHash(123456789), expectedHash)
	}
	if reloaded.TelegramOffset("test-token") != 8 || reloaded.TelegramOffset("another-token") != 0 {
		t.Fatal("Telegram polling cursor did not survive restart or token change")
	}
	if _, err := reloaded.Answer(q.ID, "Answered", false); err != nil {
		t.Fatal(err)
	}
	settings.DailyLimit = 2
	if _, err := reloaded.UpdateTelegramSettings(settings); err != nil {
		t.Fatal(err)
	}
	if reloaded.TelegramSettings().AskEnabledAt != s.TelegramSettings().AskEnabledAt {
		t.Fatal("editing Telegram settings reset the ask enable time")
	}
	_, outcome, err = reloaded.AddTelegramQuestion(123456789, 8, "test-token", "Second question")
	if err != nil || outcome != TelegramQuestionCreated {
		t.Fatalf("after answer: %s %v", outcome, err)
	}
	settings.DailyLimit = 3
	settings.AllowMultiplePending = true
	if _, err := reloaded.UpdateTelegramSettings(settings); err != nil {
		t.Fatal(err)
	}
	_, outcome, err = reloaded.AddTelegramQuestion(123456789, 9, "test-token", "Third question")
	if err != nil || outcome != TelegramQuestionCreated {
		t.Fatalf("multiple pending enabled: %s %v", outcome, err)
	}
	if _, err := os.Stat("data/telegram_identity.key"); !os.IsNotExist(err) {
		t.Fatalf("unexpected Telegram identity key file: %v", err)
	}
}

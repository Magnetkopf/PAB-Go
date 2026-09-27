package jsonstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
)

func TestLoadSettingsMigratesLegacyTelegramConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	legacy := `{"telegram_bot_token":"token","telegram_user_id":"1234"}`
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}

	settings, err := loadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if !settings.TelegramEnabled {
		t.Fatal("legacy Telegram configuration was not enabled")
	}

	persisted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved domain.Settings
	if err := json.Unmarshal(persisted, &saved); err != nil {
		t.Fatal(err)
	}
	if !saved.TelegramEnabled {
		t.Fatal("migrated telegram_enabled value was not persisted")
	}
}

func TestLoadSettingsKeepsExplicitlyDisabledTelegramConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	configured := `{"telegram_enabled":false,"telegram_bot_token":"token","telegram_user_id":"1234"}`
	if err := os.WriteFile(path, []byte(configured), 0600); err != nil {
		t.Fatal(err)
	}

	settings, err := loadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if settings.TelegramEnabled {
		t.Fatal("explicitly disabled Telegram configuration was enabled")
	}
}

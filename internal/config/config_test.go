package config

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLoadOrInitCreatesCompleteSecrets(t *testing.T) {
	chdirTemp(t)
	var output bytes.Buffer
	cfg, err := LoadOrInitFrom(bytes.NewBufferString("alice\nsuper-secret\n"), &output)
	if err != nil {
		t.Fatalf("LoadOrInitFrom() error = %v", err)
	}
	if cfg.Username != "alice" || !cfg.complete() {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte("super-secret")); err != nil {
		t.Fatalf("password hash does not verify: %v", err)
	}
	if len(cfg.SessionSecret) != sha256.Size*2 {
		t.Fatalf("session secret length = %d, want %d", len(cfg.SessionSecret), sha256.Size*2)
	}
	if _, err := os.Stat(SecretsPath); err != nil {
		t.Fatalf("secrets file not created: %v", err)
	}
}

func TestLoadOrInitReplacesIncompleteSecrets(t *testing.T) {
	chdirTemp(t)
	if err := os.MkdirAll(DataDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SecretsPath, []byte(`{"username":"old"}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadOrInitFrom(bytes.NewBufferString("new\nnew-password\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("LoadOrInitFrom() error = %v", err)
	}
	if cfg.Username != "new" {
		t.Fatalf("username = %q, want new", cfg.Username)
	}
}

func TestUpdateCredentialsPersistsPasswordAndSessionSecret(t *testing.T) {
	chdirTemp(t)
	cfg, err := LoadOrInitFrom(bytes.NewBufferString("admin\nold-password\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("new-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	secret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	updated, err := UpdateCredentials(cfg, string(passwordHash), secret)
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("new-password")); err != nil {
		t.Fatalf("updated password does not verify: %v", err)
	}
	loaded, err := LoadOrInitFrom(bytes.NewBuffer(nil), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SessionSecret != secret {
		t.Fatalf("session secret = %q, want %q", loaded.SessionSecret, secret)
	}
}

func chdirTemp(t *testing.T) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(".", DataDir)); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

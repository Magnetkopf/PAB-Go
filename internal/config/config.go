// Package config owns the small set of local files needed to start PAB-Go.
package config

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	DataDir       = "data"
	SecretsPath   = DataDir + "/secrets.json"
	SettingsPath  = DataDir + "/settings.json"
	QuestionsPath = DataDir + "/questions.json"
	UploadDir     = DataDir + "/upload"
)

type Config struct {
	Username      string `json:"username"`
	PasswordHash  string `json:"password_hash"`
	SessionSecret string `json:"session_secret"`
}

// LoadOrInit loads administrator credentials. A missing, malformed, or
// incomplete secrets file is replaced through the interactive CLI setup.
func LoadOrInit() (Config, error) { return LoadOrInitFrom(os.Stdin, os.Stdout) }

func LoadOrInitFrom(in io.Reader, out io.Writer) (Config, error) {
	b, err := os.ReadFile(SecretsPath)
	if err == nil {
		var cfg Config
		if json.Unmarshal(b, &cfg) == nil && cfg.complete() {
			return cfg, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read %s: %w", SecretsPath, err)
	}

	fmt.Fprintln(out, "Welcome to PAB-Go. Let's set admin info")
	reader := bufio.NewReader(in)
	username, err := promptRequired(reader, out, "username: ")
	if err != nil {
		return Config{}, err
	}
	password, err := promptRequired(reader, out, "password: ")
	if err != nil {
		return Config{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Config{}, fmt.Errorf("hash password: %w", err)
	}
	secret, err := newSessionSecret()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{Username: username, PasswordHash: string(passwordHash), SessionSecret: secret}
	if err := writeJSON(SecretsPath, cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) complete() bool {
	if strings.TrimSpace(c.Username) == "" || strings.TrimSpace(c.PasswordHash) == "" || len(c.SessionSecret) != sha256.Size*2 {
		return false
	}
	if _, err := hex.DecodeString(c.SessionSecret); err != nil {
		return false
	}
	_, err := bcrypt.Cost([]byte(c.PasswordHash))
	return err == nil
}

func promptRequired(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	for {
		fmt.Fprint(out, label)
		value, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("read %s: %w", strings.TrimSuffix(label, ": "), err)
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		if errors.Is(err, io.EOF) {
			return "", errors.New("initialization cancelled: value is required")
		}
		fmt.Fprintln(out, "Value cannot be empty. Please try again.")
	}
}

func newSessionSecret() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate session secret: %w", err)
	}
	randomString := base64.RawURLEncoding.EncodeToString(random)
	sum := sha256.Sum256([]byte(randomString))
	return hex.EncodeToString(sum[:]), nil
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".pab-*.json")
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

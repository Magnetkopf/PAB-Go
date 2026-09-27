package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/domain"
	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type fakeTelegramNotifier struct {
	texts     []string
	questions []domain.Question
	images    []string
	err       error
}

func (f *fakeTelegramNotifier) SendText(_ context.Context, _ domain.TelegramSettings, text string) error {
	f.texts = append(f.texts, text)
	return f.err
}

func (f *fakeTelegramNotifier) SendQuestion(_ context.Context, _ domain.TelegramSettings, question domain.Question, imagePath string) error {
	f.questions = append(f.questions, question)
	f.images = append(f.images, imagePath)
	return f.err
}

type questionTestStore struct {
	settings domain.Settings
	added    []domain.Question
}

func (s *questionTestStore) Settings() domain.Settings { return s.settings }
func (s *questionTestStore) UpdateSettings(v domain.Settings) (domain.Settings, error) {
	s.settings = v
	return v, nil
}
func (s *questionTestStore) CaptchaSettings() domain.CaptchaSettings {
	return domain.CaptchaSettings{Enabled: s.settings.CaptchaEnabled, Algorithm: s.settings.CaptchaAlgorithm, Cost: s.settings.CaptchaCost}
}
func (s *questionTestStore) UpdateCaptchaSettings(v domain.CaptchaSettings) (domain.CaptchaSettings, error) {
	s.settings.CaptchaEnabled, s.settings.CaptchaAlgorithm, s.settings.CaptchaCost = v.Enabled, v.Algorithm, v.Cost
	return v, nil
}
func (s *questionTestStore) TelegramSettings() domain.TelegramSettings {
	return domain.TelegramSettings{Enabled: s.settings.TelegramEnabled, BotToken: s.settings.TelegramBotToken, UserID: s.settings.TelegramUserID}
}
func (s *questionTestStore) UpdateTelegramSettings(v domain.TelegramSettings) (domain.TelegramSettings, error) {
	s.settings.TelegramEnabled, s.settings.TelegramBotToken, s.settings.TelegramUserID = v.Enabled, v.BotToken, v.UserID
	return v, nil
}
func (s *questionTestStore) AddQuestion(nickname, content, imageFilename string) (domain.Question, error) {
	q := domain.Question{ID: "question-1", Nickname: nickname, Content: content, ImageFilename: imageFilename}
	s.added = append(s.added, q)
	return q, nil
}
func (s *questionTestStore) Questions(string) []domain.Question { return nil }
func (s *questionTestStore) Search(string) []domain.Question    { return nil }
func (s *questionTestStore) Answer(string, string, bool) (domain.Question, error) {
	return domain.Question{}, os.ErrNotExist
}
func (s *questionTestStore) CreateSession(string, time.Time) (domain.Session, error) {
	return domain.Session{}, nil
}
func (s *questionTestStore) SessionByTokenHash(string) (domain.Session, bool) {
	return domain.Session{}, false
}
func (s *questionTestStore) Sessions() []domain.Session { return nil }
func (s *questionTestStore) RevokeSession(string) error { return os.ErrNotExist }
func (s *questionTestStore) RevokeAllSessions() error   { return nil }

func TestQuestionImageUploadStoresHashAndServesIt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chdirHandlerTemp(t)
	store := &questionTestStore{settings: domain.Settings{MaxUploadKB: 1}}
	router := gin.New()
	New(store, Config{}).Register(router.Group("/api"))

	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, '\r', 'I', 'H', 'D', 'R'}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("nickname", "Ada")
	_ = writer.WriteField("content", "Hello world")
	part, err := writer.CreateFormFile("image", "not-a-png.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(png); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/questions", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(store.added) != 1 {
		t.Fatalf("added = %d, want 1", len(store.added))
	}
	wantHash := sha256.Sum256(png)
	wantFilename := hex.EncodeToString(wantHash[:])
	if store.added[0].ImageFilename != wantFilename {
		t.Fatalf("filename = %q, want %q", store.added[0].ImageFilename, wantFilename)
	}
	stored, err := os.ReadFile(config.UploadDir + "/" + wantFilename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, png) {
		t.Fatal("stored file differs from upload")
	}

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/images/"+wantFilename, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("image status = %d", response.Code)
	}
	if response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
	if got, _ := io.ReadAll(response.Result().Body); !bytes.Equal(got, png) {
		t.Fatal("served file differs from upload")
	}
}

func TestQuestionImageUploadRejectsNonImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chdirHandlerTemp(t)
	store := &questionTestStore{settings: domain.Settings{MaxUploadKB: 1}}
	router := gin.New()
	New(store, Config{}).Register(router.Group("/api"))
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("content", "Hello world")
	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("not an image")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/questions", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(store.added) != 0 {
		t.Fatal("question was recorded for a non-image attachment")
	}
}

func TestNewQuestionSendsTelegramNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &questionTestStore{settings: domain.Settings{
		MaxUploadKB: 1, TelegramEnabled: true, TelegramBotToken: "bot-token", TelegramUserID: "1234",
	}}
	notifier := &fakeTelegramNotifier{}
	api := New(store, Config{})
	api.telegram = notifier
	router := gin.New()
	api.Register(router.Group("/api"))

	request := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"nickname":"Ada","content":"Hello world"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(notifier.questions) != 1 || notifier.questions[0].Content != "Hello world" {
		t.Fatalf("notifications = %+v, want one question", notifier.questions)
	}
	if notifier.images[0] != "" {
		t.Fatalf("image path = %q, want empty", notifier.images[0])
	}
}

func TestDisabledTelegramDoesNotSendNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &questionTestStore{settings: domain.Settings{
		MaxUploadKB: 1, TelegramBotToken: "bot-token", TelegramUserID: "1234",
	}}
	notifier := &fakeTelegramNotifier{}
	api := New(store, Config{})
	api.telegram = notifier
	router := gin.New()
	api.Register(router.Group("/api"))
	request := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"content":"Hello world"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || len(notifier.questions) != 0 {
		t.Fatalf("status = %d, notifications = %d", response.Code, len(notifier.questions))
	}
}

func TestPublicSettingsDoNotExposeTelegramCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &questionTestStore{settings: domain.Settings{
		SiteName: "AskBox", TelegramBotToken: "secret-token", TelegramUserID: "1234",
	}}
	router := gin.New()
	New(store, Config{}).Register(router.Group("/api"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/settings", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("secret-token")) || bytes.Contains(response.Body.Bytes(), []byte("1234")) {
		t.Fatalf("public settings exposed Telegram credentials: %s", response.Body.String())
	}
}

func TestTelegramTestSendsHelloWorld(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &questionTestStore{settings: domain.Settings{TelegramBotToken: "bot-token", TelegramUserID: "1234"}}
	notifier := &fakeTelegramNotifier{}
	api := New(store, Config{})
	api.telegram = notifier
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/telegram/test", nil)

	api.testTelegram(context)

	if response.Code != http.StatusOK || len(notifier.texts) != 1 || notifier.texts[0] != "helloworld" {
		t.Fatalf("status = %d, messages = %#v, body = %s", response.Code, notifier.texts, response.Body.String())
	}
}

func TestTelegramCannotBeEnabledWithoutBothCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &questionTestStore{}
	api := New(store, Config{})
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPut, "/api/admin/telegram/settings", bytes.NewBufferString(`{"telegram_enabled":true,"telegram_bot_token":"bot-token","telegram_user_id":""}`))
	context.Request.Header.Set("Content-Type", "application/json")

	api.updateTelegramSettings(context)

	if response.Code != http.StatusBadRequest || store.settings.TelegramEnabled {
		t.Fatalf("status = %d, settings = %+v, body = %s", response.Code, store.settings, response.Body.String())
	}
}

func chdirHandlerTemp(t *testing.T) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCaptchaProtectsQuestionsAndAdministratorLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settings := domain.Settings{CaptchaEnabled: true, CaptchaAlgorithm: "PBKDF2/SHA-256", CaptchaCost: 1, MaxUploadKB: 1}
	store := &questionTestStore{settings: settings}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	New(store, Config{Username: "admin", PasswordHash: string(passwordHash), SessionSecret: "captcha-test-secret"}).Register(router.Group("/api"))

	noCaptcha := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"content":"Hello world"}`))
	noCaptcha.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, noCaptcha)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("question without CAPTCHA status = %d, body = %s", response.Code, response.Body.String())
	}

	questionPayload := solveCaptcha(t, router, captchaActionQuestion)
	response = httptest.NewRecorder()
	question := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"content":"Hello world","altcha":"`+questionPayload+`"}`))
	question.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, question)
	if response.Code != http.StatusCreated {
		t.Fatalf("verified question status = %d, body = %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	replay := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"content":"Hello again","altcha":"`+questionPayload+`"}`))
	replay.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, replay)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("replayed CAPTCHA status = %d, body = %s", response.Code, response.Body.String())
	}

	loginPayload := solveCaptcha(t, router, captchaActionLogin)
	response = httptest.NewRecorder()
	login := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBufferString(`{"username":"admin","password":"secret","altcha":"`+loginPayload+`"}`))
	login.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, login)
	if response.Code != http.StatusOK {
		t.Fatalf("verified login status = %d, body = %s", response.Code, response.Body.String())
	}
}

func solveCaptcha(t *testing.T, router http.Handler, action string) string {
	t.Helper()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/captcha/challenge/"+action, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("challenge status = %d, body = %s", response.Code, response.Body.String())
	}
	var challenge altcha.Challenge
	if err := json.Unmarshal(response.Body.Bytes(), &challenge); err != nil {
		t.Fatal(err)
	}
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{Challenge: challenge, DeriveKey: altcha.DeriveKeyPBKDF2()})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(altcha.Payload{Challenge: challenge, Solution: *solution})
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(payload)
}

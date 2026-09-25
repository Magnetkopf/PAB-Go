package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/Magnetkopf/PAB-Go/internal/domain"
	"github.com/gin-gonic/gin"
)

type questionTestStore struct {
	settings domain.Settings
	added    []domain.Question
}

func (s *questionTestStore) Settings() domain.Settings { return s.settings }
func (s *questionTestStore) UpdateSettings(v domain.Settings) (domain.Settings, error) {
	s.settings = v
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

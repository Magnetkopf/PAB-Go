package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
	jsonstore "github.com/Magnetkopf/PAB-Go/internal/store/json"
)

type botTestStore struct {
	settings domain.TelegramSettings
	added    int
	content  string
}

func (s *botTestStore) TelegramSettings() domain.TelegramSettings { return s.settings }
func (s *botTestStore) TelegramOffset(string) int64               { return 0 }
func (s *botTestStore) RecordTelegramUpdate(string, int64) error  { return nil }
func (s *botTestStore) AddTelegramQuestion(userID, updateID int64, token, content string) (domain.Question, jsonstore.TelegramQuestionResult, error) {
	s.added++
	s.content = content
	return domain.Question{ID: "q1", Content: content, Status: "pending"}, jsonstore.TelegramQuestionCreated, nil
}

func TestBotTreatsIncomingHTMLAsPlainText(t *testing.T) {
	const payload = `<img src=x onerror=alert(1)>`
	var sent []struct{ chatID, text string }
	client := NewClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		switch r.URL.Path {
		case "/bottest/getUpdates":
			_, _ = recorder.Write([]byte(`{"ok":true,"result":[{"update_id":7,"message":{"text":"<img src=x onerror=alert(1)>","from":{"id":123},"chat":{"id":123,"type":"private"}}}]}`))
		case "/bottest/sendMessage":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Has("parse_mode") || r.Form.Has("entities") {
				t.Fatalf("Telegram formatting enabled: %v", r.Form)
			}
			sent = append(sent, struct{ chatID, text string }{r.FormValue("chat_id"), r.FormValue("text")})
			_, _ = recorder.Write([]byte(`{"ok":true,"result":{}}`))
		default:
			http.NotFound(recorder, r)
		}
		return recorder.Result(), nil
	})})
	client.baseURL = "https://fake.invalid"
	updates, err := client.getUpdates(context.Background(), "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	store := &botTestStore{}
	settings := domain.TelegramSettings{BotToken: "test", AskEnabled: true, PushEnabled: true, UserID: "999"}
	for _, item := range updates {
		if err := client.handleUpdate(context.Background(), store, settings, item); err != nil {
			t.Fatal(err)
		}
	}
	if store.added != 1 || store.content != payload {
		t.Fatalf("stored question: added=%d content=%q", store.added, store.content)
	}
	if len(sent) != 2 || sent[0].chatID != "999" || sent[0].text != "New question from Anonymous\n\n"+payload || sent[1].chatID != "123" || sent[1].text != "Your question has been added to the inbox." {
		t.Fatalf("sent messages = %#v", sent)
	}
}

func TestBotReceivesPrivateTextOnly(t *testing.T) {
	var mu sync.Mutex
	sent := []string{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest/getUpdates":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":7,"message":{"text":"A private question","from":{"id":123,"is_bot":false},"chat":{"id":123,"type":"private"}}},{"update_id":8,"message":{"text":"A group question","from":{"id":123,"is_bot":false},"chat":{"id":456,"type":"group"}}}]}`))
		case "/bottest/sendMessage":
			_ = r.ParseForm()
			mu.Lock()
			sent = append(sent, r.FormValue("chat_id")+":"+r.FormValue("text"))
			mu.Unlock()
			_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
		default:
			http.NotFound(w, r)
		}
	})
	client := NewClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		return recorder.Result(), nil
	})})
	client.baseURL = "https://fake.invalid"
	updates, err := client.getUpdates(context.Background(), "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	store := &botTestStore{settings: domain.TelegramSettings{BotToken: "test", AskEnabled: true}}
	for _, item := range updates {
		if err := client.handleUpdate(context.Background(), store, store.settings, item); err != nil {
			t.Fatal(err)
		}
	}
	if store.added != 1 {
		t.Fatalf("created questions = %d", store.added)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(sent) != 1 || sent[0] != "123:Your question has been added to the inbox." {
		t.Fatalf("messages = %#v", sent)
	}
}

func TestBotIgnoresMessageSentBeforeAskWasEnabled(t *testing.T) {
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("old message must not trigger a Telegram reply")
		return nil, nil
	})})
	store := &botTestStore{}
	var item update
	if err := json.Unmarshal([]byte(`{"update_id":9,"message":{"text":"Old question","date":100,"from":{"id":123},"chat":{"id":123,"type":"private"}}}`), &item); err != nil {
		t.Fatal(err)
	}
	if err := client.handleUpdate(context.Background(), store, domain.TelegramSettings{AskEnabledAt: 101}, item); err != nil {
		t.Fatal(err)
	}
	if store.added != 0 {
		t.Fatal("old question was accepted")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

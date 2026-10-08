package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
)

const defaultBaseURL = "https://api.telegram.org"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{httpClient: httpClient, baseURL: defaultBaseURL}
}

func (c *Client) SendText(ctx context.Context, settings domain.TelegramSettings, text string) error {
	return c.SendChatText(ctx, settings.BotToken, settings.UserID, text)
}

func (c *Client) SendChatText(ctx context.Context, token, chatID, text string) error {
	values := url.Values{"chat_id": {chatID}, "text": {text}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL(token, "sendMessage"), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func (c *Client) SendQuestion(ctx context.Context, settings domain.TelegramSettings, question domain.Question, imagePath string) error {
	message := questionMessage(question)
	if imagePath == "" {
		return c.SendText(ctx, settings, truncateRunes(message, 4096))
	}

	image, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("open question image: %w", err)
	}
	defer image.Close()
	contentType, extension, err := imageType(image)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("chat_id", settings.UserID); err != nil {
		return err
	}
	if err := writer.WriteField("caption", truncateRunes(message, 1024)); err != nil {
		return err
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name": "photo", "filename": filepath.Base(imagePath) + extension,
	}))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, image); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL(settings.BotToken, "sendPhoto"), &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(req)
}

func imageType(image *os.File) (contentType, extension string, err error) {
	buffer := make([]byte, 512)
	n, readErr := image.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		return "", "", fmt.Errorf("read question image: %w", readErr)
	}
	if _, err := image.Seek(0, io.SeekStart); err != nil {
		return "", "", fmt.Errorf("rewind question image: %w", err)
	}
	contentType = http.DetectContentType(buffer[:n])
	switch contentType {
	case "image/png":
		extension = ".png"
	case "image/jpeg":
		extension = ".jpg"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		extension = ".image"
	}
	return contentType, extension, nil
}

func (c *Client) methodURL(token, method string) string {
	return strings.TrimRight(c.baseURL, "/") + "/bot" + url.PathEscape(token) + "/" + method
}

func (c *Client) do(req *http.Request) error {
	response, err := c.httpClient.Do(req)
	if err != nil {
		// net/http errors include the full request URL, which contains the bot
		// token. Keep that credential out of logs and API responses.
		return fmt.Errorf("Telegram request failed")
	}
	defer response.Body.Close()
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(&result); err != nil {
		return fmt.Errorf("decode Telegram response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || !result.OK {
		if result.Description == "" {
			result.Description = response.Status
		}
		return fmt.Errorf("Telegram API: %s", result.Description)
	}
	return nil
}

func questionMessage(question domain.Question) string {
	name := strings.TrimSpace(question.Nickname)
	if name == "" {
		name = "Anonymous"
	}
	return fmt.Sprintf("New question from %s\n\n%s", name, question.Content)
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit-1]) + "…"
}

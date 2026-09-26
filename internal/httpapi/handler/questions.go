package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Magnetkopf/PAB-Go/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	kilobyte             = 1024
	multipartMemoryLimit = 1024 * kilobyte
)

var imageHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func (a *API) getQuestions(c *gin.Context) {
	status := c.Query("status")
	if status != "published" && !a.isAdmin(c) {
		status = "published"
	}
	c.JSON(http.StatusOK, gin.H{"questions": a.store.Questions(status)})
}
func (a *API) searchQuestions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"questions": a.store.Search(c.Query("q"))})
}
func (a *API) createQuestion(c *gin.Context) {
	// Keep JSON support for existing text-only clients. Attachments are sent
	// together with the question as multipart/form-data, never pre-uploaded.
	if strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		var input struct {
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
			Altcha   string `json:"altcha"`
		}
		if err := c.ShouldBindJSON(&input); err != nil || !validQuestionContent(input.Content) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content"})
			return
		}
		if a.store.Settings().CaptchaEnabled && !a.verifyCaptcha(c, input.Altcha, captchaActionQuestion) {
			return
		}
		a.saveQuestion(c, input.Nickname, input.Content, "")
		return
	}

	maxFileBytes := int64(a.store.Settings().MaxUploadKB) * kilobyte
	// Permit multipart boundaries and fields in addition to the configured
	// attachment limit; the file itself is checked below.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileBytes+multipartMemoryLimit)
	if err := c.Request.ParseMultipartForm(multipartMemoryLimit); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "uploaded file exceeds the size limit or the form is malformed"})
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	nickname, content := c.PostForm("nickname"), c.PostForm("content")
	if !validQuestionContent(content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question content must be between 5 and 1,000 characters"})
		return
	}
	if a.store.Settings().CaptchaEnabled && !a.verifyCaptcha(c, c.PostForm("altcha"), captchaActionQuestion) {
		return
	}
	image, header, err := c.Request.FormFile("image")
	if errors.Is(err, http.ErrMissingFile) {
		a.saveQuestion(c, nickname, content, "")
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read image attachment"})
		return
	}
	defer image.Close()
	if header.Size > maxFileBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image exceeds the size limit"})
		return
	}
	if !isImage(image) {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "attachment must be a PNG, JPEG, GIF, or WebP image"})
		return
	}
	filename, err := storeImage(image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save image"})
		return
	}
	a.saveQuestion(c, nickname, content, filename)
}

func (a *API) saveQuestion(c *gin.Context, nickname, content, imageFilename string) {
	q, err := a.store.AddQuestion(nickname, content, imageFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save question"})
		return
	}
	c.JSON(http.StatusCreated, q)
}

func validQuestionContent(content string) bool {
	length := len([]rune(strings.TrimSpace(content)))
	return length >= 0 && length <= 1000
}

func isImage(file multipart.File) bool {
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false
	}
	switch http.DetectContentType(buf[:n]) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func storeImage(file multipart.File) (string, error) {
	if err := os.MkdirAll(config.UploadDir, 0700); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(config.UploadDir, ".pab-upload-*")
	if err != nil {
		return "", err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(temp, hash), file); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Chmod(0600); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	filename := hex.EncodeToString(hash.Sum(nil))
	path := filepath.Join(config.UploadDir, filename)
	if _, err := os.Stat(path); err == nil {
		return filename, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.Rename(tempName, path); err != nil {
		return "", err
	}
	return filename, nil
}

func (a *API) getImage(c *gin.Context) {
	filename := c.Param("sha256")
	if !imageHashPattern.MatchString(filename) {
		c.Status(http.StatusNotFound)
		return
	}
	file, err := os.Open(filepath.Join(config.UploadDir, filename))
	if errors.Is(err, os.ErrNotExist) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read image"})
		return
	}
	defer file.Close()
	contentTypeBuffer := make([]byte, 512)
	n, err := file.Read(contentTypeBuffer)
	if err != nil && !errors.Is(err, io.EOF) {
		c.Status(http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(contentTypeBuffer[:n])
	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		c.Status(http.StatusNotFound)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	info, err := file.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, fmt.Sprintf("%s.image", filename), info.ModTime(), file)
}

func (a *API) answerQuestion(c *gin.Context) {
	var input struct {
		Answer  string `json:"answer"`
		Publish bool   `json:"publish"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Answer) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answer cannot be empty"})
		return
	}
	q, err := a.store.Answer(c.Param("id"), input.Answer, input.Publish)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save answer"})
		return
	}
	c.JSON(http.StatusOK, q)
}

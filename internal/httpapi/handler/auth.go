package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (a *API) getSession(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"authenticated": a.isAdmin(c)}) }
func (a *API) login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&input) != nil || !hmac.Equal([]byte(input.Username), []byte(a.config.Username)) || bcrypt.CompareHashAndPassword([]byte(a.config.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}
	c.SetCookie("askbox_session", signSession(a.config.SessionSecret), 7*24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
func (a *API) logout(c *gin.Context) {
	c.SetCookie("askbox_session", "", -1, "/", "", false, true)
	c.Status(http.StatusNoContent)
}
func (a *API) requireAdmin(c *gin.Context) {
	if !a.isAdmin(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "administrator sign-in required"})
		return
	}
	c.Next()
}
func (a *API) isAdmin(c *gin.Context) bool {
	cookie, err := c.Cookie("askbox_session")
	return err == nil && validSession(cookie, a.config.SessionSecret)
}
func signSession(secret string) string {
	expires := time.Now().Add(7 * 24 * time.Hour).Unix()
	payload := fmt.Sprintf("admin:%d", expires)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func validSession(raw, secret string) bool {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	given, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	if !hmac.Equal(given, mac.Sum(nil)) {
		return false
	}
	var expires int64
	if _, err := fmt.Sscanf(string(payload), "admin:%d", &expires); err != nil {
		return false
	}
	return time.Now().Unix() < expires
}

package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "askbox_session"
	sessionLifetime   = 30 * 24 * time.Hour
	sessionContextKey = "admin_session"
)

func (a *API) getSession(c *gin.Context) {
	session, ok := a.activeSession(c)
	response := gin.H{"authenticated": ok}
	if ok {
		response["session_id"] = session.ID
	}
	c.JSON(http.StatusOK, response)
}

func (a *API) login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	a.credentials.RLock()
	username, passwordHash, sessionSecret := a.config.Username, a.config.PasswordHash, a.config.SessionSecret
	a.credentials.RUnlock()
	if c.ShouldBindJSON(&input) != nil || !hmac.Equal([]byte(input.Username), []byte(username)) || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}
	token, err := newSessionToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create session"})
		return
	}
	if _, err := a.store.CreateSession(sessionTokenHash(token, sessionSecret), time.Now().Add(sessionLifetime)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create session"})
		return
	}
	c.SetCookie(sessionCookieName, token, int(sessionLifetime.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) resetPassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if c.ShouldBindJSON(&input) != nil || input.CurrentPassword == "" || input.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password and new password are required"})
		return
	}

	// Keep verification and replacement together so concurrent reset requests
	// cannot both validate against the same old password.
	a.credentials.Lock()
	defer a.credentials.Unlock()
	if bcrypt.CompareHashAndPassword([]byte(a.config.PasswordHash), []byte(input.CurrentPassword)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update password"})
		return
	}
	sessionSecret, err := newSessionSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update password"})
		return
	}
	if a.config.UpdateCredentials == nil || a.config.UpdateCredentials(string(passwordHash), sessionSecret) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update password"})
		return
	}

	// Rotating the secret invalidates every existing cookie before the session
	// store is cleared, so there is no window in which an old login survives.
	a.config.PasswordHash = string(passwordHash)
	a.config.SessionSecret = sessionSecret
	if err := a.store.RevokeAllSessions(); err != nil {
		a.clearSessionCookie(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password updated but unable to revoke sessions"})
		return
	}
	a.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}
func (a *API) logout(c *gin.Context) {
	session := c.MustGet(sessionContextKey).(domain.Session)
	if err := a.store.RevokeSession(session.ID); err != nil && !errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to log out"})
		return
	}
	a.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (a *API) listSessions(c *gin.Context) {
	current := c.MustGet(sessionContextKey).(domain.Session)
	sessions := a.store.Sessions()
	result := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, gin.H{
			"id":         session.ID,
			"created_at": session.CreatedAt,
			"expires_at": session.ExpiresAt,
			"current":    session.ID == current.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sessions": result})
}

func (a *API) revokeSession(c *gin.Context) {
	current := c.MustGet(sessionContextKey).(domain.Session)
	id := c.Param("id")
	if err := a.store.RevokeSession(id); err != nil {
		// DELETE is intentionally idempotent: a duplicated browser request or a
		// session revoked by another device has already reached the requested
		// secure state, so it is still a successful revoke.
		if !errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to revoke session"})
			return
		}
	}
	if id == current.ID {
		a.clearSessionCookie(c)
	}
	c.Status(http.StatusNoContent)
}

func (a *API) requireAdmin(c *gin.Context) {
	session, ok := a.activeSession(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "administrator sign-in required"})
		return
	}
	c.Set(sessionContextKey, session)
	c.Next()
}
func (a *API) activeSession(c *gin.Context) (domain.Session, bool) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil || cookie == "" {
		return domain.Session{}, false
	}
	a.credentials.RLock()
	sessionSecret := a.config.SessionSecret
	a.credentials.RUnlock()
	return a.store.SessionByTokenHash(sessionTokenHash(cookie, sessionSecret))
}

func (a *API) isAdmin(c *gin.Context) bool {
	_, ok := a.activeSession(c)
	return ok
}

func (a *API) clearSessionCookie(c *gin.Context) {
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
}
func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func newSessionSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func sessionTokenHash(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

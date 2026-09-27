package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
	"github.com/gin-gonic/gin"
)

func (a *API) getSettings(c *gin.Context) {
	settings := a.store.Settings()
	settings.TelegramEnabled = false
	settings.TelegramBotToken = ""
	settings.TelegramUserID = ""
	c.JSON(http.StatusOK, settings)
}

func (a *API) updateSettings(c *gin.Context) {
	var next domain.Settings
	if c.ShouldBindJSON(&next) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings format"})
		return
	}
	saved, err := a.store.UpdateSettings(next)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save settings"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (a *API) getCaptchaSettings(c *gin.Context) { c.JSON(http.StatusOK, a.store.CaptchaSettings()) }

func (a *API) updateCaptchaSettings(c *gin.Context) {
	var next domain.CaptchaSettings
	if c.ShouldBindJSON(&next) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CAPTCHA settings format"})
		return
	}
	saved, err := a.store.UpdateCaptchaSettings(next)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save CAPTCHA settings"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (a *API) getTelegramSettings(c *gin.Context) {
	c.JSON(http.StatusOK, a.store.TelegramSettings())
}

func (a *API) updateTelegramSettings(c *gin.Context) {
	var next domain.TelegramSettings
	if c.ShouldBindJSON(&next) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Telegram settings format"})
		return
	}
	next.BotToken = strings.TrimSpace(next.BotToken)
	next.UserID = strings.TrimSpace(next.UserID)
	if next.Enabled && (next.BotToken == "" || next.UserID == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telegram bot token and user ID are required before Telegram can be enabled"})
		return
	}
	saved, err := a.store.UpdateTelegramSettings(next)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save Telegram settings"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (a *API) testTelegram(c *gin.Context) {
	settings := a.store.TelegramSettings()
	if settings.BotToken == "" || settings.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telegram is not configured"})
		return
	}
	if err := a.telegram.SendText(c.Request.Context(), settings, "helloworld"); err != nil {
		log.Printf("send Telegram test message: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "unable to send Telegram test message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": true})
}

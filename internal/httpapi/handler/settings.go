package handler

import (
	"net/http"

	"github.com/Magnetkopf/PAB-Go/internal/domain"
	"github.com/gin-gonic/gin"
)

func (a *API) getSettings(c *gin.Context) { c.JSON(http.StatusOK, a.store.Settings()) }

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

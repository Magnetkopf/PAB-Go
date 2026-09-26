package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"sync"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/gin-gonic/gin"
)

const (
	captchaActionQuestion = "question"
	captchaActionLogin    = "login"
	captchaLifetime       = 5 * time.Minute
)

type captchaState struct {
	mu   sync.Mutex
	used map[string]time.Time
}

// captchaChallenge gives the ALTCHA widget a new, signed proof-of-work
// challenge. Challenges are action-bound and expire quickly.
func (a *API) captchaChallenge(c *gin.Context) {
	action := c.Param("action")
	if !validCaptchaAction(action) || !a.store.Settings().CaptchaEnabled {
		c.Status(http.StatusNotFound)
		return
	}
	settings := a.store.Settings()
	counter, err := captchaCounter()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create verification challenge"})
		return
	}
	expiresAt := time.Now().Add(captchaLifetime)
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           settings.CaptchaAlgorithm,
		Cost:                settings.CaptchaCost,
		Counter:             &counter,
		DeriveKey:           captchaDeriveKey(settings.CaptchaAlgorithm),
		ExpiresAt:           &expiresAt,
		HMACSignatureSecret: a.captchaSecret(),
		Data:                map[string]interface{}{"action": action},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create verification challenge"})
		return
	}
	c.JSON(http.StatusOK, challenge)
}

func (a *API) verifyCaptcha(c *gin.Context, encodedPayload, action string) bool {
	if encodedPayload == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "complete the CAPTCHA to continue"})
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(encodedPayload)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CAPTCHA verification failed"})
		return false
	}
	var payload altcha.Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CAPTCHA verification failed"})
		return false
	}
	settings := a.store.Settings()
	params := payload.Challenge.Parameters
	if params.Algorithm != settings.CaptchaAlgorithm || params.Cost != settings.CaptchaCost || !validCaptchaAction(action) || captchaAction(params.Data) != action {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CAPTCHA verification failed"})
		return false
	}
	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge:           payload.Challenge,
		Solution:            payload.Solution,
		DeriveKey:           captchaDeriveKey(settings.CaptchaAlgorithm),
		HMACSignatureSecret: a.captchaSecret(),
	})
	if err != nil || result.Expired || !result.Verified || payload.Challenge.Signature == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CAPTCHA verification failed or expired"})
		return false
	}
	if !a.consumeCaptcha(payload.Challenge.Signature, time.Unix(params.ExpiresAt, 0)) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CAPTCHA verification has already been used"})
		return false
	}
	return true
}

func (a *API) captchaSecret() string {
	a.credentials.RLock()
	defer a.credentials.RUnlock()
	return a.config.SessionSecret
}

func (a *API) consumeCaptcha(signature string, expiresAt time.Time) bool {
	a.captcha.mu.Lock()
	defer a.captcha.mu.Unlock()
	now := time.Now()
	for key, expiry := range a.captcha.used {
		if !expiry.After(now) {
			delete(a.captcha.used, key)
		}
	}
	if _, used := a.captcha.used[signature]; used {
		return false
	}
	a.captcha.used[signature] = expiresAt
	return true
}

func captchaCounter() (int, error) {
	// This follows ALTCHA's deterministic challenge pattern: the client finds
	// one counter from a bounded random range instead of a fixed solution.
	n, err := rand.Int(rand.Reader, big.NewInt(9001))
	if err != nil {
		return 0, err
	}
	return 1000 + int(n.Int64()), nil
}

func captchaDeriveKey(algorithm string) altcha.DeriveKeyFunc {
	if len(algorithm) >= len("PBKDF2/") && algorithm[:len("PBKDF2/")] == "PBKDF2/" {
		return altcha.DeriveKeyPBKDF2()
	}
	return altcha.DeriveKeySHA()
}

func validCaptchaAction(action string) bool {
	return action == captchaActionQuestion || action == captchaActionLogin
}

func captchaAction(data map[string]interface{}) string {
	action, _ := data["action"].(string)
	return action
}

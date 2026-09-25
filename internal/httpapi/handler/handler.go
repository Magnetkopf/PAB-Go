package handler

import (
	"github.com/Magnetkopf/PAB-Go/internal/domain"
	"github.com/gin-gonic/gin"
)

type Store interface {
	Settings() domain.Settings
	UpdateSettings(domain.Settings) (domain.Settings, error)
	AddQuestion(string, string, string) (domain.Question, error)
	Questions(string) []domain.Question
	Search(string) []domain.Question
	Answer(string, string, bool) (domain.Question, error)
}

type Config struct {
	Username, PasswordHash, SessionSecret string
}
type API struct {
	store  Store
	config Config
}

func New(store Store, config Config) *API { return &API{store: store, config: config} }

func (a *API) Register(api *gin.RouterGroup) {
	api.GET("/settings", a.getSettings)
	api.GET("/questions", a.getQuestions)
	api.GET("/questions/search", a.searchQuestions)
	api.POST("/questions", a.createQuestion)
	api.GET("/images/:sha256", a.getImage)
	api.GET("/admin/session", a.getSession)
	api.POST("/admin/login", a.login)
	admin := api.Group("/admin", a.requireAdmin)
	admin.POST("/logout", a.logout)
	admin.PUT("/settings", a.updateSettings)
	admin.POST("/questions/:id/answer", a.answerQuestion)
}

package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:assets
var assets embed.FS

func Serve(router *gin.Engine) {
	files, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	router.NoRoute(func(c *gin.Context) {
		name := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if name == "." {
			name = "index.html"
		}
		contents, err := fs.ReadFile(files, name)
		if err != nil {
			name = "index.html"
			contents, err = fs.ReadFile(files, name)
		}
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		contentType := mime.TypeByExtension(path.Ext(name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Data(http.StatusOK, contentType, contents)
	})
}

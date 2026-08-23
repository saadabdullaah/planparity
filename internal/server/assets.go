package server

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed assets
var embeddedAssets embed.FS

func (server *Server) assets() http.Handler {
	root := "assets"
	if _, err := fs.Stat(embeddedAssets, "assets/ui/index.html"); err == nil {
		root = "assets/ui"
	}
	assets, _ := fs.Sub(embeddedAssets, root)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested == "." || requested == "" {
			requested = "index.html"
		}
		content, err := fs.ReadFile(assets, requested)
		if err != nil {
			requested = "index.html"
			content, err = fs.ReadFile(assets, requested)
		}
		if err != nil {
			http.NotFound(writer, request)
			return
		}
		if contentType := mime.TypeByExtension(path.Ext(requested)); contentType != "" {
			writer.Header().Set("Content-Type", contentType)
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(content)
	})
}

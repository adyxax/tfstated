package webui

import (
	"html/template"
	"net/http"
)

var indexTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/index.html"))

func handleIndexGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")

		render(w, indexTemplates, http.StatusOK, nil)
	})
}

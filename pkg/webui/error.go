package webui

import (
	"html/template"
	"net/http"
)

var errorTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/error.html"))

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	type ErrorData struct {
		Page       *Page
		Err        error
		Status     int
		StatusText string
	}
	render(w, errorTemplates, status, &ErrorData{
		Page:       &Page{Title: "Error", Section: "error"},
		Err:        err,
		Status:     status,
		StatusText: http.StatusText(status),
	})
}

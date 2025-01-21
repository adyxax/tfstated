package webui

import (
	"html/template"
	"net/http"
)

var errorTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/error.html"))

func errorResponse(w http.ResponseWriter, status int, err error) {
	type ErrorData struct {
		Err        error
		Status     int
		StatusText string
	}
	render(w, errorTemplates, status, &ErrorData{
		Err:        err,
		Status:     status,
		StatusText: http.StatusText(status),
	})
}
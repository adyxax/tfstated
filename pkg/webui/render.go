package webui

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

func render(w http.ResponseWriter, t *template.Template, status int, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(
			"%s: failed to execute template: %+v",
			http.StatusText(http.StatusInternalServerError),
			err)))
	} else {
		w.WriteHeader(status)
		_, _ = buf.WriteTo(w)
	}
}

package webui

import (
	"fmt"
	"net/http"
)

func handleIndexGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/states", http.StatusFound)
		} else {
			errorResponse(w, http.StatusNotFound, fmt.Errorf("Page not found"))
		}
	})
}

package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type Page struct {
	Section string
	Session *model.Session
	Title   string
}

func makePage(r *http.Request, page *Page) *Page {
	page.Session = r.Context().Value(model.SessionContextKey{}).(*model.Session)
	return page
}

func handleIndexGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/states", http.StatusFound)
		} else {
			errorResponse(w, r, http.StatusNotFound, fmt.Errorf("Page not found"))
		}
	})
}

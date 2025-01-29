package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type Page struct {
	LightMode bool
	Precedent string
	Section   string
	Title     string
}

func makePage(r *http.Request, page *Page) *Page {
	settings := r.Context().Value(model.SettingsContextKey{}).(*model.Settings)
	page.LightMode = settings.LightMode
	return page
}

func handleIndexGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/states", http.StatusFound)
		} else {
			errorResponse(w, http.StatusNotFound, fmt.Errorf("Page not found"))
		}
	})
}

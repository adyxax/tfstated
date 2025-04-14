package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

type Page struct {
	AccountId uuid.UUID
	IsAdmin   bool
	LightMode bool
	Section   string
	Title     string
}

func makePage(r *http.Request, page *Page) *Page {
	account := r.Context().Value(model.AccountContextKey{}).(*model.Account)
	page.AccountId = account.Id
	page.IsAdmin = account.IsAdmin
	settings := r.Context().Value(model.SettingsContextKey{}).(*model.Settings)
	page.LightMode = settings.LightMode
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

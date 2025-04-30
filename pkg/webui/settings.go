package webui

import (
	"context"
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type SettingsPage struct {
	Page     *Page
	Settings *model.Settings
}

var settingsTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/settings.html"))

func handleSettingsGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render(w, settingsTemplates, http.StatusOK, SettingsPage{
			Page: makePage(r, &Page{Title: "Settings", Section: "settings"}),
		})
	})
}

func handleSettingsPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		darkMode := r.FormValue("dark-mode")
		settings := model.Settings{
			LightMode: darkMode != "1",
		}
		session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
		session.Data.Settings = &settings
		err := db.SaveAccountSettings(session.Data.Account, &settings)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		ctx := context.WithValue(r.Context(), model.SessionContextKey{}, session)
		render(w, settingsTemplates, http.StatusOK, SettingsPage{
			Page:     makePage(r.WithContext(ctx), &Page{Title: "Settings", Section: "settings"}),
			Settings: &settings,
		})
	})
}

package webui

import (
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
		settings := r.Context().Value(model.SettingsContextKey{}).(*model.Settings)
		render(w, settingsTemplates, http.StatusOK, SettingsPage{
			Page:     makePage(r, &Page{Title: "Settings", Section: "settings"}),
			Settings: settings,
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
		account := r.Context().Value(model.AccountContextKey{}).(*model.Account)
		err := db.SaveAccountSettings(account, &settings)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		render(w, settingsTemplates, http.StatusOK, SettingsPage{
			Page: &Page{
				LightMode: settings.LightMode,
				Title:     "Settings",
				Section:   "settings",
			},
			Settings: &settings,
		})
	})
}

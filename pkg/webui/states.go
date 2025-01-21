package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var statesTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/states.html"))

func handleStatesGET(db *database.DB) http.Handler {
	type StatesData struct {
		States []model.State
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")

		states, err := db.LoadStatesByPath()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		render(w, statesTemplates, http.StatusOK, StatesData{
			States: states,
		})
	})
}

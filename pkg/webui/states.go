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
		Page   *Page
		States []model.State
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		states, err := db.LoadStates()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		render(w, statesTemplates, http.StatusOK, StatesData{
			Page:   makePage(r, &Page{Title: "States", Section: "states"}),
			States: states,
		})
	})
}

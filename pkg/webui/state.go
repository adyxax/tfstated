package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

var stateTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/state.html"))

func handleStateGET(db *database.DB) http.Handler {
	type StatesData struct {
		Page      *Page
		State     *model.State
		Usernames map[string]string
		Versions  []model.Version
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var stateId uuid.UUID
		if err := stateId.Parse(r.PathValue("id")); err != nil {
			errorResponse(w, http.StatusBadRequest, err)
			return
		}
		state, err := db.LoadStateById(stateId)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		versions, err := db.LoadVersionsByState(state)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		usernames, err := db.LoadAccountUsernames()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		render(w, stateTemplate, http.StatusOK, StatesData{
			Page: makePage(r, &Page{
				Precedent: "/states",
				Section:   "states",
				Title:     state.Path,
			}),
			State:     state,
			Usernames: usernames,
			Versions:  versions,
		})
	})
}

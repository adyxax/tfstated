package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

type StatesIdPage struct {
	Page          *Page
	Path          string
	PathError     bool
	PathDuplicate bool
	State         *model.State
	Usernames     map[string]string
	Versions      []model.Version
}

var statesIdTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/statesId.html"))

func handleStatesIdGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var stateId uuid.UUID
		if err := stateId.Parse(r.PathValue("id")); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		state, err := db.LoadStateById(stateId)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		versions, err := db.LoadVersionsByState(state)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		usernames, err := db.LoadAccountUsernames()
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		render(w, statesIdTemplate, http.StatusOK, StatesIdPage{
			Page: makePage(r, &Page{
				Section: "states",
				Title:   state.Path,
			}),
			State:     state,
			Usernames: usernames,
			Versions:  versions,
		})
	})
}

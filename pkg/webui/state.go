package webui

import (
	"html/template"
	"net/http"
	"strconv"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var stateTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/state.html"))

func handleStateGET(db *database.DB) http.Handler {
	type StatesData struct {
		Page      *Page
		State     *model.State
		Usernames map[int]string
		Versions  []model.Version
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateIdStr := r.PathValue("id")
		stateId, err := strconv.Atoi(stateIdStr)
		if err != nil {
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

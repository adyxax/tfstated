package webui

import (
	"html/template"
	"net/http"
	"net/url"
	"path"

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

func handleStatesIdPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
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
		action := r.FormValue("action")
		switch action {
		case "delete":
			errorResponse(w, r, http.StatusNotImplemented, err)
		case "edit":
			statePath := r.FormValue("path")
			parsedStatePath, err := url.Parse(statePath)
			if err != nil || path.Clean(parsedStatePath.Path) != statePath || statePath[0] != '/' {
				render(w, statesIdTemplate, http.StatusBadRequest, StatesIdPage{
					Page:      makePage(r, &Page{Title: state.Path, Section: "states"}),
					Path:      statePath,
					PathError: true,
					State:     state,
					Usernames: usernames,
					Versions:  versions,
				})
				return
			}
			state.Path = statePath
			success, err := db.SaveState(state)
			if err != nil {
				errorResponse(w, r, http.StatusInternalServerError, err)
				return
			}
			if !success {
				render(w, statesIdTemplate, http.StatusBadRequest, StatesIdPage{
					Page:          makePage(r, &Page{Title: state.Path, Section: "states"}),
					Path:          statePath,
					PathDuplicate: true,
					State:         state,
					Usernames:     usernames,
					Versions:      versions,
				})
				return
			}
		case "unlock":
			if err := db.ForceUnlock(state); err != nil {
				errorResponse(w, r, http.StatusInternalServerError, err)
				return
			}
			state.Lock = nil
		default:
			errorResponse(w, r, http.StatusBadRequest, err)
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

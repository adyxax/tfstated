package webui

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type StatesPage struct {
	Page          *Page
	Path          string
	PathError     bool
	PathDuplicate bool
	States        []model.State
}

var statesTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/states.html"))

func handleStatesGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		states, err := db.LoadStates()
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		render(w, statesTemplates, http.StatusOK, StatesPage{
			Page:   makePage(r, &Page{Title: "States", Section: "states"}),
			States: states,
		})
	})
}

func handleStatesPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		states, err := db.LoadStates()
		// file upload limit of 20MB
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		defer file.Close()
		statePath := r.FormValue("path")
		parsedStatePath, err := url.Parse(statePath)
		if err != nil || path.Clean(parsedStatePath.Path) != statePath || statePath[0] != '/' {
			render(w, statesTemplates, http.StatusBadRequest, StatesPage{
				Page:      makePage(r, &Page{Title: "States", Section: "states"}),
				Path:      statePath,
				PathError: true,
				States:    states,
			})
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			errorResponse(w, r, http.StatusBadRequest, fmt.Errorf("failed to read uploaded file: %w", err))
			return
		}
		fileType := http.DetectContentType(data)
		if fileType != "text/plain; charset=utf-8" {
			errorResponse(w, r, http.StatusBadRequest, fmt.Errorf("invalid file type: expected \"text/plain; charset=utf-8\" but got \"%s\"", fileType))
			return
		}
		account := r.Context().Value(model.AccountContextKey{}).(*model.Account)
		version, err := db.CreateState(statePath, account.Id, data)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if version == nil {
			render(w, statesTemplates, http.StatusBadRequest, StatesPage{
				Page:          makePage(r, &Page{Title: "State", Section: "states"}),
				Path:          statePath,
				PathDuplicate: true,
				States:        states,
			})
			return
		}
		destination := path.Join("/versions", version.Id.String())
		http.Redirect(w, r, destination, http.StatusFound)
	})
}

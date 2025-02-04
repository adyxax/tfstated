package webui

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type StatesNewPage struct {
	Page          *Page
	fileError     bool
	Path          string
	PathDuplicate bool
	PathError     bool
}

var statesNewTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/states_new.html"))

func handleStatesNewGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render(w, statesNewTemplates, http.StatusOK, StatesNewPage{
			Page: makePage(r, &Page{Title: "New State", Section: "states"}),
		})
	})
}

func handleStatesNewPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// file upload limit of 20MB
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			errorResponse(w, http.StatusBadRequest, err)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err)
			return
		}
		defer file.Close()
		statePath := r.FormValue("path")
		parsedStatePath, err := url.Parse(statePath)
		if err != nil || path.Clean(parsedStatePath.Path) != statePath || statePath[0] != '/' {
			render(w, statesNewTemplates, http.StatusBadRequest, StatesNewPage{
				Page:      makePage(r, &Page{Title: "New State", Section: "states"}),
				Path:      statePath,
				PathError: true,
			})
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to read uploaded file: %w", err))
			return
		}
		fileType := http.DetectContentType(data)
		if fileType != "text/plain; charset=utf-8" {
			errorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid file type: expected \"text/plain; charset=utf-8\" but got \"%s\"", fileType))
			return
		}
		account := r.Context().Value(model.AccountContextKey{}).(*model.Account)
		version, err := db.CreateState(statePath, account.Id, data)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		if version == nil {
			render(w, statesNewTemplates, http.StatusBadRequest, StatesNewPage{
				Page:          makePage(r, &Page{Title: "New State", Section: "states"}),
				Path:          statePath,
				PathDuplicate: true,
			})
			return
		}
		destination := path.Join("/version", strconv.Itoa(version.Id))
		http.Redirect(w, r, destination, http.StatusFound)
	})
}

package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

var versionsTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/versions.html"))

func handleVersionsGET(db *database.DB) http.Handler {
	type VersionsData struct {
		Page        *Page
		Account     *model.Account
		State       *model.State
		Version     *model.Version
		VersionData string
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var versionId uuid.UUID
		if err := versionId.Parse(r.PathValue("id")); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		version, err := db.LoadVersionById(versionId)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if version == nil {
			errorResponse(w, r, http.StatusNotFound, err)
			return
		}
		state, err := db.LoadStateById(version.StateId)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		account, err := db.LoadAccountById(&version.AccountId)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		versionData := string(version.Data[:])
		render(w, versionsTemplate, http.StatusOK, VersionsData{
			Page: makePage(r, &Page{
				Section: "states",
				Title:   state.Path,
			}),
			Account:     account,
			State:       state,
			Version:     version,
			VersionData: versionData,
		})
	})
}

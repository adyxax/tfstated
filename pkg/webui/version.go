package webui

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var versionTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/version.html"))

func handleVersionGET(db *database.DB) http.Handler {
	type VersionsData struct {
		Page
		Account     *model.Account
		State       *model.State
		Version     *model.Version
		VersionData string
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionIdStr := r.PathValue("id")
		versionId, err := strconv.Atoi(versionIdStr)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, err)
			return
		}
		version, err := db.LoadVersionById(versionId)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		state, err := db.LoadStateById(version.StateId)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		account, err := db.LoadAccountById(version.AccountId)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		versionData := string(version.Data[:])
		render(w, versionTemplate, http.StatusOK, VersionsData{
			Page: Page{
				Precedent: fmt.Sprintf("/state/%d", state.Id),
				Section:   "versions",
				Title:     state.Path,
			},
			Account:     account,
			State:       state,
			Version:     version,
			VersionData: versionData,
		})
	})
}

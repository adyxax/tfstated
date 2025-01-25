package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var logoutTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/logout.html"))

func handleLogoutGET(db *database.DB) http.Handler {
	type logoutPage struct {
		Page
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(model.SessionContextKey{})
		err := db.DeleteSession(session.(*model.Session))
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		unsetSesssionCookie(w)
		render(w, logoutTemplate, http.StatusOK, logoutPage{
			Page: Page{Title: "Logout", Section: "login"},
		})
	})
}

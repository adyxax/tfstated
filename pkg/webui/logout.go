package webui

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var logoutTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/logout.html"))

func handleLogoutGET(db *database.DB) http.Handler {
	type logoutPage struct {
		Page *Page
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
		sessionId, session, err := db.MigrateSession(session, nil)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to migrate session: %w", err))
			return
		}
		setSessionCookie(w, sessionId)
		ctx := context.WithValue(r.Context(), model.SessionContextKey{}, session)
		render(w, logoutTemplate, http.StatusOK, logoutPage{
			Page: makePage(r.WithContext(ctx), &Page{Title: "Logout", Section: "login"}),
		})
	})
}

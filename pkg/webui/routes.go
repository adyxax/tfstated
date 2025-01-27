package webui

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	requireSession := sessionsMiddleware(db)
	requireLogin := loginMiddleware(db, requireSession)
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("GET /login", requireSession(handleLoginGET()))
	mux.Handle("POST /login", requireSession(handleLoginPOST(db)))
	mux.Handle("GET /logout", requireLogin(handleLogoutGET(db)))
	mux.Handle("GET /states", requireLogin(handleStatesGET(db)))
	mux.Handle("GET /state/{id}", requireLogin(handleStateGET(db)))
	mux.Handle("GET /static/", cache(http.FileServer(http.FS(staticFS))))
	mux.Handle("GET /version/{id}", requireLogin(handleVersionGET(db)))
	mux.Handle("GET /", requireLogin(handleIndexGET()))
}

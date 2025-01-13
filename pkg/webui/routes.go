package webui

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	session := sessionsMiddleware(db)
	requireLogin := loginMiddleware(db)
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("GET /login", session(handleLoginGET()))
	mux.Handle("POST /login", session(handleLoginPOST(db)))
	mux.Handle("GET /logout", session(requireLogin(handleLogoutGET(db))))
	mux.Handle("GET /static/", cache(http.FileServer(http.FS(staticFS))))
	mux.Handle("GET /", session(requireLogin(handleIndexGET())))
}

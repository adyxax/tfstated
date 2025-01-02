package backend

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/middlewares/basic_auth"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	mux.Handle("GET /healthz", handleHealthz())

	basicAuth := basic_auth.Middleware(db)
	mux.Handle("DELETE /", basicAuth(handleDelete(db)))
	mux.Handle("GET /", basicAuth(handleGet(db)))
	mux.Handle("LOCK /", basicAuth(handleLock(db)))
	mux.Handle("POST /", basicAuth(handlePost(db)))
	mux.Handle("UNLOCK /", basicAuth(handleUnlock(db)))
}

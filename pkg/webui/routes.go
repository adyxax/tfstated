package webui

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	mux.Handle("GET /healthz", handleHealthz())
}

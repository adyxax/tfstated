package main

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	mux.Handle("GET /healthz", handleHealthz())

	mux.Handle("DELETE /", handleDelete(db))
	mux.Handle("GET /", handleGet(db))
	mux.Handle("POST /", handlePost(db))
}

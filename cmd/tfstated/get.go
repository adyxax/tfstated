package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func handleGet(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{\"msg\": \"No state path provided, cannot GET /\"}"))
			return
		}

		if data, err := db.GetState(r.URL.Path); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _ = w.Write([]byte(fmt.Sprintf("{\"msg\": \"%+v\"}", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		}
	})
}

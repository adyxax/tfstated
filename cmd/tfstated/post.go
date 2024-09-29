package main

import (
	"fmt"
	"io"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func handlePost(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{\"msg\": \"No state path provided, cannot POST /\"}"))
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("{\"msg\": \"failed to read request body: %+v\"}", err)))
			return
		}
		if err := db.SetState(r.URL.Path, data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("{\"msg\": \"%+v\"}", err)))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

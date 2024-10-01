package main

import (
	"fmt"
	"io"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func handlePost(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = errorResponse(w, http.StatusBadRequest,
				fmt.Errorf("no state path provided, cannot POST /"),
			)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			_ = errorResponse(w, http.StatusBadRequest, err)
			return
		}
		if err := db.SetState(r.URL.Path, data); err != nil {
			_ = errorResponse(w, http.StatusInternalServerError, err)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

package main

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
)

func handleGet(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")

		if r.URL.Path == "/" {
			helpers.ErrorResponse(w, http.StatusBadRequest,
				fmt.Errorf("no state path provided, cannot GET /"))
			return
		}

		if data, err := db.GetState(r.URL.Path); err != nil {
			helpers.ErrorResponse(w, http.StatusInternalServerError, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		}
	})
}

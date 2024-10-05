package main

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func handleDelete(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = errorResponse(w, http.StatusBadRequest,
				fmt.Errorf("no state path provided, cannot DELETE /"))
			return
		}

		if success, err := db.DeleteState(r.URL.Path); err != nil {
			_ = errorResponse(w, http.StatusInternalServerError, err)
		} else if success {
			w.WriteHeader(http.StatusOK)
		} else {
			_ = errorResponse(w, http.StatusNotFound,
				fmt.Errorf("state path not found: %s", r.URL.Path))
		}
	})
}

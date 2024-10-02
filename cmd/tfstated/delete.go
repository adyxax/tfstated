package main

import (
	"database/sql"
	"errors"
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

		if err := db.DeleteState(r.URL.Path); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = errorResponse(w, http.StatusNotFound,
					fmt.Errorf("state path not found: %s", r.URL.Path))
			} else {
				_ = errorResponse(w, http.StatusInternalServerError, err)
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

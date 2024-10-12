package main

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func handleUnlock(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = encode(w, http.StatusBadRequest,
				fmt.Errorf("no state path provided, cannot LOCK /"))
			return
		}

		var lock lockRequest
		if err := decode(r, &lock); err != nil {
			_ = encode(w, http.StatusBadRequest, err)
			return
		}
		if success, err := db.Unlock(r.URL.Path, &lock); err != nil {
			_ = errorResponse(w, http.StatusInternalServerError, err)
		} else if success {
			w.WriteHeader(http.StatusOK)
		} else {
			_ = encode(w, http.StatusConflict, lock)
		}
	})
}

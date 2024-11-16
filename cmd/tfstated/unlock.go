package main

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
)

func handleUnlock(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = helpers.Encode(w, http.StatusBadRequest,
				fmt.Errorf("no state path provided, cannot LOCK /"))
			return
		}

		var lock lockRequest
		if err := helpers.Decode(r, &lock); err != nil {
			_ = helpers.Encode(w, http.StatusBadRequest, err)
			return
		}
		if success, err := db.Unlock(r.URL.Path, &lock); err != nil {
			helpers.ErrorResponse(w, http.StatusInternalServerError, err)
		} else if success {
			w.WriteHeader(http.StatusOK)
		} else {
			_ = helpers.Encode(w, http.StatusConflict, lock)
		}
	})
}

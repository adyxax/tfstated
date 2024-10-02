package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

type lockRequest struct {
	Created   time.Time `json:"Created"`
	ID        string    `json:"ID"`
	Info      string    `json:"Info"`
	Operation string    `json:"Operation"`
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Who       string    `json:"Who"`
}

func handleLock(db *database.DB) http.Handler {
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
		if success, err := db.SetLockOrGetExistingLock(r.URL.Path, &lock); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = encode(w, http.StatusNotFound,
					fmt.Errorf("state path not found: %s", r.URL.Path))
			} else {
				_ = errorResponse(w, http.StatusInternalServerError, err)
			}
		} else if success {
			w.WriteHeader(http.StatusOK)
		} else {
			_ = encode(w, http.StatusConflict, lock)
		}
	})
}

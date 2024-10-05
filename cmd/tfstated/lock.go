package main

import (
	"fmt"
	"net/http"
	"regexp"
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

var (
	validID = regexp.MustCompile("[a-f0-9]{8}-(?:[a-f0-9]{4}-){3}[a-f0-9]{12}")
)

func (l *lockRequest) valid() []error {
	err := make([]error, 0)
	if !validID.MatchString(l.ID) {
		err = append(err, fmt.Errorf("invalid ID"))
	}
	return err
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
		if errs := lock.valid(); len(errs) > 0 {
			_ = encode(w, http.StatusBadRequest,
				fmt.Errorf("invalid lock: %+v", errs))
			return
		}
		if success, err := db.SetLockOrGetExistingLock(r.URL.Path, &lock); err != nil {
			_ = errorResponse(w, http.StatusInternalServerError, err)
		} else if success {
			w.WriteHeader(http.StatusOK)
		} else {
			_ = encode(w, http.StatusConflict, lock)
		}
	})
}

package basic_auth

import (
	"context"
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

func Middleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="tfstated", charset="UTF-8"`)
				helpers.ErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
				return
			}
			account, err := db.LoadAccountByUsername(username)
			if err != nil {
				helpers.ErrorResponse(w, http.StatusInternalServerError, err)
				return
			}
			if account == nil || !account.CheckPassword(password) {
				helpers.ErrorResponse(w, http.StatusForbidden, fmt.Errorf("Forbidden"))
				return
			}
			if err := db.TouchAccount(account); err != nil {
				helpers.ErrorResponse(w, http.StatusInternalServerError, err)
				return
			}
			ctx := context.WithValue(r.Context(), model.AccountContextKey{}, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

func adminMiddleware(db *database.DB, requireLogin func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			account := r.Context().Value(model.AccountContextKey{})
			if account == nil {
				// this could happen if the account was deleted in the short
				// time between retrieving the session and here
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			if !account.(*model.Account).IsAdmin {
				errorResponse(w, http.StatusForbidden, fmt.Errorf("Only administrators can perform this request."))
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

func adminMiddleware(requireLogin func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
			if !session.Data.Account.IsAdmin {
				errorResponse(w, r, http.StatusForbidden, fmt.Errorf("Only administrators can perform this request."))
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

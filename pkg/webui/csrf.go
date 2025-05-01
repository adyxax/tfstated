package webui

import (
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

func verifyCSRFToken(w http.ResponseWriter, r *http.Request) bool {
	session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
	tokenStr := r.FormValue("csrf_token")
	if tokenStr == "" {
		tokenStr = r.Header.Get("X-XSRF-Token")
	}
	var token uuid.UUID
	if err := token.Parse(tokenStr); err != nil {
		errorResponse(w, r, http.StatusBadRequest,
			fmt.Errorf("failed to parse csrf token: %w", err))
		return false
	}

	if !token.Equal(session.Data.CsrfToken) {
		errorResponse(w, r, http.StatusForbidden,
			fmt.Errorf("invalid csrf token"))
		return false
	}
	return true
}

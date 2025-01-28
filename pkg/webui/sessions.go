package webui

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

const cookieName = "tfstated"

func sessionsMiddleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				errorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to get request cookie \"%s\": %w", cookieName, err))
				return
			}
			if err == nil {
				if len(cookie.Value) != 36 {
					unsetSesssionCookie(w)
				} else {
					session, err := db.LoadSessionById(cookie.Value)
					if err != nil {
						errorResponse(w, http.StatusInternalServerError, err)
						return
					}
					if session == nil {
						unsetSesssionCookie(w)
					} else if !session.IsExpired() {
						if err := db.TouchSession(cookie.Value); err != nil {
							errorResponse(w, http.StatusInternalServerError, err)
							return
						}
						ctx := context.WithValue(r.Context(), model.SessionContextKey{}, session)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func unsetSesssionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Quoted:   false,
		Path:     "/",
		MaxAge:   0, // remove invalid cookie
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

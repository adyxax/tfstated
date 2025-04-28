package webui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
				errorResponse(w, r, http.StatusInternalServerError,
					fmt.Errorf("failed to get session cookie from request: %w", err))
				return
			}
			if err == nil {
				if len(cookie.Value) == 43 {
					session, err := db.LoadSessionById(cookie.Value)
					if err != nil {
						errorResponse(w, r, http.StatusInternalServerError,
							fmt.Errorf("failed to load session by ID: %w", err))
						return
					}
					if session != nil {
						if session.IsExpired() {
							if err := db.DeleteSession(session); err != nil {
								errorResponse(w, r, http.StatusInternalServerError,
									fmt.Errorf("failed to delete session: %w", err))
								return
							}
						} else {
							ctx := context.WithValue(r.Context(), model.SessionContextKey{}, session)
							var settings model.Settings
							if err := json.Unmarshal(session.Settings, &settings); err != nil {
								slog.Error("failed to unmarshal session settings", "err", err)
							}
							ctx = context.WithValue(ctx, model.SettingsContextKey{}, &settings)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}
			sessionId, err := db.CreateSession(nil, nil)
			if err != nil {
				errorResponse(w, r, http.StatusInternalServerError,
					fmt.Errorf("failed to create session: %w", err))
				return
			}
			setSessionCookie(w, sessionId)
			var settings model.Settings
			ctx := context.WithValue(r.Context(), model.SettingsContextKey{}, &settings)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setSessionCookie(w http.ResponseWriter, sessionId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    sessionId,
		Quoted:   false,
		Path:     "/",
		MaxAge:   12 * 3600, // 12 hours sessions
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

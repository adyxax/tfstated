package basic_auth

import (
	"context"
	"net/http"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

func Middleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="tfstated", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			account, err := db.LoadAccountByUsername(username)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if password != account.Password {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			now := time.Now().UTC()
			_, err = db.Exec(`UPDATE accounts SET last_login = ? WHERE id = ?`, now.Unix(), account.Id)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), model.AccountContextKey{}, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

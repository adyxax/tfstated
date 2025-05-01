package webui

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var loginTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/login.html"))

var validUsername = regexp.MustCompile(`^[a-zA-Z]\w*$`)

type loginPage struct {
	Page      *Page
	Forbidden bool
	Username  string
}

func handleLoginGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")

		session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
		if session.Data.Account != nil {
			http.Redirect(w, r, "/states", http.StatusFound)
			return
		}

		render(w, loginTemplate, http.StatusOK, loginPage{
			Page: makePage(r, &Page{Title: "Login", Section: "login"}),
		})
	})
}

func handleLoginPOST(db *database.DB) http.Handler {
	renderForbidden := func(w http.ResponseWriter, r *http.Request, username string) {
		render(w, loginTemplate, http.StatusForbidden, loginPage{
			Page:      makePage(r, &Page{Title: "Login", Section: "login"}),
			Forbidden: true,
			Username:  username,
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest,
				fmt.Errorf("failed to parse form: %w", err))
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			errorResponse(w, r, http.StatusBadRequest, fmt.Errorf("Invalid username or password"))
			return
		}
		if ok := validUsername.MatchString(username); !ok {
			renderForbidden(w, r, username)
			return
		}
		account, err := db.LoadAccountByUsername(username)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to load account by username %s: %w", username, err))
			return
		}
		if account == nil || !account.CheckPassword(password) {
			renderForbidden(w, r, username)
			return
		}
		if err := db.TouchAccount(account); err != nil {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to touch account %s: %w", username, err))
			return
		}
		session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
		sessionId, session, err := db.MigrateSession(session, account)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to migrate session: %w", err))
			return
		}
		setSessionCookie(w, sessionId)
		ctx := context.WithValue(r.Context(), model.SessionContextKey{}, session)
		if err := db.DeleteExpiredSessions(); err != nil {
			slog.Error("failed to delete expired sessions after user login", "err", err, "accountId", account.Id)
		}
		http.Redirect(w, r.WithContext(ctx), "/", http.StatusFound)
	})
}

func loginMiddleware(requireSession func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store, no-cache")
			session := r.Context().Value(model.SessionContextKey{}).(*model.Session)
			if session.Data.Account == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

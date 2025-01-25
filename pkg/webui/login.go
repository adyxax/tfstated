package webui

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

var loginTemplate = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/login.html"))

type loginPage struct {
	Page
	Forbidden bool
	Username  string
}

func handleLoginGET() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache")

		session := r.Context().Value(model.SessionContextKey{})
		if session != nil {
			http.Redirect(w, r, "/states", http.StatusFound)
			return
		}

		render(w, loginTemplate, http.StatusOK, loginPage{
			Page: Page{Title: "Login", Section: "login"},
		})
	})
}

func handleLoginPOST(db *database.DB) http.Handler {
	var validUsername = regexp.MustCompile(`^[a-zA-Z]\w*$`)
	renderForbidden := func(w http.ResponseWriter, username string) {
		render(w, loginTemplate, http.StatusForbidden, loginPage{
			Page:      Page{Title: "Login", Section: "login"},
			Forbidden: true,
			Username:  username,
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, http.StatusBadRequest, err)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" { // the webui cannot issue this
			errorResponse(w, http.StatusBadRequest, fmt.Errorf("Forbidden"))
			return
		}
		if ok := validUsername.MatchString(username); !ok {
			renderForbidden(w, username)
			return
		}
		account, err := db.LoadAccountByUsername(username)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		if account == nil || !account.CheckPassword(password) {
			renderForbidden(w, username)
			return
		}
		if err := db.TouchAccount(account); err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		sessionId, err := db.CreateSession(account)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    sessionId,
			Quoted:   false,
			Path:     "/",
			MaxAge:   8 * 3600, // 1 hour sessions
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
		})
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func loginMiddleware(db *database.DB, requireSession func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store, no-cache")
			session := r.Context().Value(model.SessionContextKey{})
			if session == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			account, err := db.LoadAccountById(session.(*model.Session).AccountId)
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}
			if account == nil {
				// this could happen if the account was deleted in the short
				// time between retrieving the session and here
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			ctx := context.WithValue(r.Context(), model.AccountContextKey{}, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

package webui

import (
	"fmt"
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

type AccountsIdResetPasswordPage struct {
	Account         *model.Account
	Page            *Page
	PasswordInvalid bool
	PasswordChanged bool
	Token           string
}

var accountsIdResetPasswordTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/accountsIdResetPassword.html"))

func processAccountsIdResetPasswordPathValues(db *database.DB, w http.ResponseWriter, r *http.Request) *model.Account {
	var accountId uuid.UUID
	if err := accountId.Parse(r.PathValue("id")); err != nil {
		return nil
	}
	var token uuid.UUID
	if err := token.Parse(r.PathValue("token")); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return nil
	}
	account, err := db.LoadAccountById(&accountId)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return nil
	}
	if account == nil || account.PasswordReset == nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return nil
	}
	if !account.PasswordReset.Equal(token) {
		errorResponse(w, r, http.StatusBadRequest, err)
		return nil
	}
	return account
}

func handleAccountsIdResetPasswordGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := processAccountsIdResetPasswordPathValues(db, w, r)
		if account == nil {
			return
		}
		render(w, accountsIdResetPasswordTemplates, http.StatusOK,
			AccountsIdResetPasswordPage{
				Account: account,
				Page:    makePage(r, &Page{Title: "Password Reset", Section: "reset"}),
				Token:   r.PathValue("token"),
			})
	})
}

func handleAccountsIdResetPasswordPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := processAccountsIdResetPasswordPathValues(db, w, r)
		if account == nil {
			return
		}
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest,
				fmt.Errorf("failed to parse form: %w", err))
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
		password := r.FormValue("password")
		if len(password) < 8 {
			errorResponse(w, r, http.StatusBadRequest, nil)
			return
		}
		account.SetPassword(password)
		success, err := db.SaveAccount(account)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to save account: %w", err))
			return
		}
		if !success {
			errorResponse(w, r, http.StatusInternalServerError,
				fmt.Errorf("failed to save account: table constraint error"))
			return
		}
		render(w, accountsIdResetPasswordTemplates, http.StatusOK,
			AccountsIdResetPasswordPage{
				Account:         account,
				Page:            makePage(r, &Page{Title: "Password Reset", Section: "reset"}),
				PasswordChanged: true,
			})
	})
}

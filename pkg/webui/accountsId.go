package webui

import (
	"fmt"
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

type AccountsIdPage struct {
	Account           *model.Account
	IsAdmin           string
	Page              *Page
	Username          string
	StatePaths        map[string]string
	UsernameDuplicate bool
	UsernameInvalid   bool
	Versions          []model.Version
}

var accountsIdTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/accountsId.html"))

func prepareAccountsIdPage(db *database.DB, w http.ResponseWriter, r *http.Request) *AccountsIdPage {
	var accountId uuid.UUID
	if err := accountId.Parse(r.PathValue("id")); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return nil
	}
	account, err := db.LoadAccountById(&accountId)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return nil
	}
	if account == nil {
		errorResponse(w, r, http.StatusNotFound, fmt.Errorf("The account Id could not be found."))
		return nil
	}
	statePaths, err := db.LoadStatePaths()
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return nil
	}
	versions, err := db.LoadVersionsByAccount(account)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return nil
	}
	isAdmin := ""
	if account.IsAdmin {
		isAdmin = "1"
	}
	return &AccountsIdPage{
		Account: account,
		IsAdmin: isAdmin,
		Page: makePage(r, &Page{
			Section: "accounts",
			Title:   account.Username,
		}),
		StatePaths: statePaths,
		Versions:   versions,
	}

}

func handleAccountsIdGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := prepareAccountsIdPage(db, w, r)
		if page != nil {
			render(w, accountsIdTemplates, http.StatusOK, page)
		}
	})
}

func handleAccountsIdPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
		page := prepareAccountsIdPage(db, w, r)
		if page == nil {
			return
		}
		action := r.FormValue("action")
		switch action {
		case "delete":
			errorResponse(w, r, http.StatusNotImplemented, nil)
			return
		case "edit":
			errorResponse(w, r, http.StatusNotImplemented, nil)
			return
		case "reset-password":
			if err := page.Account.ResetPassword(); err != nil {
				errorResponse(w, r, http.StatusNotImplemented,
					fmt.Errorf("failed to reset password: %w", err))
				return
			}
			if err := db.SaveAccount(page.Account); err != nil {
				errorResponse(w, r, http.StatusInternalServerError,
					fmt.Errorf("failed to save account: %w", err))
				return
			}
			if err := db.DeleteSessions(page.Account); err != nil {
				errorResponse(w, r, http.StatusInternalServerError,
					fmt.Errorf("failed to save account: %w", err))
				return
			}
		default:
			errorResponse(w, r, http.StatusBadRequest, nil)
			return
		}
		render(w, accountsIdTemplates, http.StatusOK, page)
	})
}

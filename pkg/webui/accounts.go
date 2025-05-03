package webui

import (
	"fmt"
	"html/template"
	"net/http"
	"path"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type AccountsPage struct {
	Accounts          []model.Account
	IsAdmin           string
	Page              *Page
	Username          string
	UsernameDuplicate bool
	UsernameInvalid   bool
}

var accountsTemplates = template.Must(template.ParseFS(htmlFS, "html/base.html", "html/accounts.html"))

func handleAccountsGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accounts, err := db.LoadAccounts()
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		render(w, accountsTemplates, http.StatusOK, AccountsPage{
			Page:     makePage(r, &Page{Title: "User Accounts", Section: "accounts"}),
			Accounts: accounts,
		})
	})
}

func handleAccountsPOST(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			errorResponse(w, r, http.StatusBadRequest,
				fmt.Errorf("failed to parse form: %w", err))
			return
		}
		if !verifyCSRFToken(w, r) {
			return
		}
		accounts, err := db.LoadAccounts()
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		accountUsername := r.FormValue("username")
		isAdmin := r.FormValue("is-admin")
		page := AccountsPage{
			Page:     makePage(r, &Page{Title: "New Account", Section: "accounts"}),
			Accounts: accounts,
			IsAdmin:  isAdmin,
			Username: accountUsername,
		}
		if ok := validUsername.MatchString(accountUsername); !ok {
			page.UsernameInvalid = true
			render(w, accountsTemplates, http.StatusBadRequest, page)
			return
		}
		account, err := db.CreateAccount(accountUsername, isAdmin == "1")
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if account == nil {
			page.UsernameDuplicate = true
			render(w, accountsTemplates, http.StatusBadRequest, page)
			return
		}
		destination := path.Join("/accounts", account.Id.String())
		http.Redirect(w, r, destination, http.StatusFound)
	})
}

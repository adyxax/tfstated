package webui

import (
	"html/template"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

type AccountsPage struct {
	ActiveTab int
	Page      *Page
	Accounts  []model.Account
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

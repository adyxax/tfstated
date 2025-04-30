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

func handleAccountsIdGET(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var accountId uuid.UUID
		if err := accountId.Parse(r.PathValue("id")); err != nil {
			errorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		account, err := db.LoadAccountById(&accountId)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if account == nil {
			errorResponse(w, r, http.StatusNotFound, fmt.Errorf("The account Id could not be found."))
			return
		}
		statePaths, err := db.LoadStatePaths()
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		versions, err := db.LoadVersionsByAccount(account)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		isAdmin := ""
		if account.IsAdmin {
			isAdmin = "1"
		}
		render(w, accountsIdTemplates, http.StatusOK, AccountsIdPage{
			Account: account,
			IsAdmin: isAdmin,
			Page: makePage(r, &Page{
				Section: "accounts",
				Title:   account.Username,
			}),
			StatePaths: statePaths,
			Versions:   versions,
		})
	})
}

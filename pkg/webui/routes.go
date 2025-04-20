package webui

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	requireSession := sessionsMiddleware(db)
	requireLogin := loginMiddleware(db, requireSession)
	requireAdmin := adminMiddleware(db, requireLogin)
	mux.Handle("GET /accounts", requireLogin(handleAccountsGET(db)))
	mux.Handle("GET /accounts/{id}", requireLogin(handleAccountsIdGET(db)))
	mux.Handle("GET /accounts/{id}/reset/{token}", handleAccountsIdResetPasswordGET(db))
	mux.Handle("POST /accounts/{id}/reset/{token}", handleAccountsIdResetPasswordPOST(db))
	mux.Handle("POST /accounts", requireAdmin(handleAccountsPOST(db)))
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("GET /login", requireSession(handleLoginGET()))
	mux.Handle("POST /login", requireSession(handleLoginPOST(db)))
	mux.Handle("GET /logout", requireLogin(handleLogoutGET(db)))
	mux.Handle("GET /settings", requireLogin(handleSettingsGET(db)))
	mux.Handle("POST /settings", requireLogin(handleSettingsPOST(db)))
	mux.Handle("GET /states", requireLogin(handleStatesGET(db)))
	mux.Handle("POST /states", requireLogin(handleStatesPOST(db)))
	mux.Handle("GET /states/{id}", requireLogin(handleStatesIdGET(db)))
	mux.Handle("GET /static/", cache(http.FileServer(http.FS(staticFS))))
	mux.Handle("GET /versions/{id}", requireLogin(handleVersionsGET(db)))
	mux.Handle("GET /", requireLogin(handleIndexGET()))
}

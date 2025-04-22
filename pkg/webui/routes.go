package webui

import (
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

func addRoutes(
	mux *http.ServeMux,
	db *database.DB,
) {
	processSession := sessionsMiddleware(db)
	requireLogin := loginMiddleware(db, processSession)
	requireAdmin := adminMiddleware(db, requireLogin)
	mux.Handle("GET /accounts", requireLogin(handleAccountsGET(db)))
	mux.Handle("GET /accounts/{id}", requireLogin(handleAccountsIdGET(db)))
	mux.Handle("GET /accounts/{id}/reset/{token}", handleAccountsIdResetPasswordGET(db))
	mux.Handle("POST /accounts/{id}/reset/{token}", handleAccountsIdResetPasswordPOST(db))
	mux.Handle("POST /accounts", requireAdmin(handleAccountsPOST(db)))
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("GET /login", processSession(handleLoginGET()))
	mux.Handle("POST /login", processSession(handleLoginPOST(db)))
	mux.Handle("GET /logout", requireLogin(handleLogoutGET(db)))
	mux.Handle("GET /settings", requireLogin(handleSettingsGET(db)))
	mux.Handle("POST /settings", requireLogin(handleSettingsPOST(db)))
	mux.Handle("GET /states", requireLogin(handleStatesGET(db)))
	mux.Handle("POST /states", requireLogin(handleStatesPOST(db)))
	mux.Handle("GET /states/{id}", requireLogin(handleStatesIdGET(db)))
	mux.Handle("POST /states/{id}", requireLogin(handleStatesIdPOST(db)))
	mux.Handle("GET /static/", cache(http.FileServer(http.FS(staticFS))))
	mux.Handle("GET /versions/{id}", requireLogin(handleVersionsGET(db)))
	mux.Handle("GET /", requireLogin(handleIndexGET()))
}

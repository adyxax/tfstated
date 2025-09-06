package webui

import (
	"context"
	"embed"
	"log/slog"
	"net"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/middlewares/logger"
)

//go:embed html/*
var htmlFS embed.FS

//go:embed static/*
var staticFS embed.FS

func Run(
	ctx context.Context,
	cancel context.CancelFunc,
	db *database.DB,
	getenv func(string) string,
) *http.Server {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		db,
	)

	host := getenv("TFSTATED_WEBUI_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := getenv("TFSTATED_WEBUI_PORT")
	if port == "" {
		port = "8081"
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: logger.Middleware(mux, false),
	}
	go func() {
		defer cancel()
		slog.Info("webui http server listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error listening and serving webui http server", "address", httpServer.Addr, "error", err)
		}
	}()

	return httpServer
}

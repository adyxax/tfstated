package backend

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/logger"
)

func Run(
	ctx context.Context,
	db *database.DB,
	//args []string,
	getenv func(string) string,
	//stdin io.Reader,
	//stdout io.Writer,
	stderr io.Writer,
) *http.Server {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		db,
	)

	host := getenv("TFSTATED_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := getenv("TFSTATED_PORT")
	if port == "" {
		port = "8080"
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: logger.Middleware(mux, false),
	}
	go func() {
		slog.Info("backend http server listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error listening and serving backend http server", "address", httpServer.Addr, "error", err)
		}
	}()

	return httpServer
}

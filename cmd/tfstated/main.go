package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/logger"
)

func run(
	ctx context.Context,
	db *database.DB,
	//args []string,
	getenv func(string) string,
	//stdin io.Reader,
	//stdout io.Writer,
	stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := db.InitAdminAccount(); err != nil {
		return err
	}

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
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_, _ = fmt.Fprintf(stderr, "error listening and serving: %+v\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			_, _ = fmt.Fprintf(stderr, "error shutting down http server: %+v\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()

	var opts *slog.HandlerOptions
	if os.Getenv("TFSTATED_DEBUG") != "" {
		opts = &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	db, err := database.NewDB(
		ctx,
		"./tfstate.db?_txlock=immediate",
		os.Getenv,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database init error: %+v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := run(
		ctx,
		db,
		//os.Args,
		os.Getenv,
		//os.Stdin,
		//os.Stdout,
		os.Stderr,
	); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

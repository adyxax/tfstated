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
	"strconv"
	"sync"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/logger"
)

type Config struct {
	Host string
	Port string
}

func run(
	ctx context.Context,
	config *Config,
	db *database.DB,
	//args []string,
	getenv func(string) string,
	//stdin io.Reader,
	//stdout io.Writer,
	stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	dataEncryptionKey := getenv("DATA_ENCRYPTION_KEY")
	if dataEncryptionKey == "" {
		return fmt.Errorf("the DATA_ENCRYPTION_KEY environment variable is not set")
	}
	if err := db.SetDataEncryptionKey(dataEncryptionKey); err != nil {
		return err
	}
	versionsHistoryLimit := getenv("VERSIONS_HISTORY_LIMIT")
	if versionsHistoryLimit != "" {
		n, err := strconv.Atoi(versionsHistoryLimit)
		if err != nil {
			return fmt.Errorf("failed to parse the VERSIONS_HISTORY_LIMIT environment variable: %w", err)
		}
		db.SetVersionsHistoryLimit(n)
	}

	mux := http.NewServeMux()
	addRoutes(
		mux,
		db,
	)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: logger.Middleware(mux, false),
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "error listening and serving: %+v\n", err)
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
			fmt.Fprintf(stderr, "error shutting down http server: %+v\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()

	var opts *slog.HandlerOptions
	if os.Getenv("TFSTATE_DEBUG") != "" {
		opts = &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	config := Config{
		Host: "0.0.0.0",
		Port: "8080",
	}

	db, err := database.NewDB(ctx, "./tfstate.db?_txlock=immediate")
	if err != nil {
		fmt.Fprintf(os.Stderr, "database init error: %+v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := run(
		ctx,
		&config,
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

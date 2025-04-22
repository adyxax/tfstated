package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/backend"
	"git.adyxax.org/adyxax/tfstated/pkg/database"
	"git.adyxax.org/adyxax/tfstated/pkg/webui"
)

func run(
	ctx context.Context,
	db *database.DB,
	getenv func(string) string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := db.InitAdminAccount(); err != nil {
		return err
	}

	backend := backend.Run(ctx, db, getenv)
	webui := webui.Run(ctx, db, getenv)

	<-ctx.Done()
	shutdownCtx := context.Background()
	shutdownCtx, shutdownCancel := context.WithTimeout(shutdownCtx, 10*time.Second)
	defer shutdownCancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := backend.Shutdown(shutdownCtx); err != nil {
			slog.Error("error shutting down backend http server", "error", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := webui.Shutdown(shutdownCtx); err != nil {
			slog.Error("error shutting down webui http server", "error", err)
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
		"./tfstated.db?_txlock=immediate",
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
		os.Getenv,
	); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

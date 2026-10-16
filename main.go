package main

import (
	"context"
	"errors"
	"fmt"
	"go-codeexplorer/repository"
	"go-codeexplorer/service"
	"go-codeexplorer/tx"
	"go-codeexplorer/web"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	log := slog.New(handler).With(slog.String("service.name", "todo-app"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("POSTGRESQL_URL")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	if autoMigrate() {
		if err := repository.Migrate(dsn, log); err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
	}

	entryRepo := repository.NewTodoEntryRepository(pool)
	txManager := tx.NewTransactionalManager(pool)
	todoService := service.NewTodoService(entryRepo, txManager)
	todoHandler := web.NewTodoHandler(todoService, log)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           web.NewRouter(todoHandler, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func autoMigrate() bool {
	v := os.Getenv("AUTO_MIGRATE")
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	return err == nil && b
}

package main

import (
	"context"
	"fmt"
	"github.com/cap79/GopherSocial/internal/apiserver"
	"github.com/cap79/GopherSocial/internal/config"
	"github.com/cap79/GopherSocial/internal/store"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	jsonHandler := slog.NewJSONHandler(w, nil)
	logger := slog.New(jsonHandler)

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(w, "%s\n", err)
	}

	db, err := store.NewPostgresDb(cfg)
	if err != nil {
		fmt.Fprintf(w, "%s\n", err)
	}

	dataStore := store.NewStore(db)

	jwtManager := apiserver.NewJwtManager(cfg)

	server := apiserver.New(cfg, logger, dataStore, jwtManager)
	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(w, "%s\n", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

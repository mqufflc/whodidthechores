package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/mqufflc/whodidthechores/internal/api"
	"github.com/mqufflc/whodidthechores/internal/config"
	"github.com/mqufflc/whodidthechores/internal/database"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

const (
	exitFail = 1
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}

func run() error {
	ctx := context.Background()

	config, err := config.New()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	pool, err := database.Connect(ctx, config.Database)
	if err != nil {
		return fmt.Errorf("database connect error: %w", err)
	}
	defer pool.Close()

	repo := repository.New(repository.NewRepositoryParams{DB: pool})

	handler := api.New(repo, config)
	http := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: handler,
	}

	fmt.Printf("Listening on :%d\n", config.Port)
	http.ListenAndServe()
	return nil
}

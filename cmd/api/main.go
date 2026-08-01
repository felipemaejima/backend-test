package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felipemaejima/backend-test/internal/config"
	httpapi "github.com/felipemaejima/backend-test/internal/http"
	"github.com/felipemaejima/backend-test/internal/repository/postgres"
	"github.com/felipemaejima/backend-test/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("aplicação encerrada com erro", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	db, err := postgres.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() {
		if err := postgres.Close(db); err != nil {
			slog.Error("falha ao fechar conexão", "error", err)
		}
	}()

	partRepository := postgres.NewPartRepository(db)

	partService := service.NewPartService(partRepository)
	restockService := service.NewRestockService(partRepository)

	app := httpapi.NewRouter(
		httpapi.NewPartHandler(partService),
		httpapi.NewRestockHandler(restockService),
		httpapi.NewHealthHandler(postgres.Pinger(db)),
	)

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("servidor iniciado", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-stop:
		slog.Info("encerrando servidor")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return app.ShutdownWithContext(ctx)
}

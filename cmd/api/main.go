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
	"github.com/felipemaejima/backend-test/internal/logger"
	"github.com/felipemaejima/backend-test/internal/repository/postgres"
	"github.com/felipemaejima/backend-test/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("aplicação encerrada com erro", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	log, closeLog, err := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		File:   cfg.Log.File,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := closeLog(); err != nil {
			slog.Error("falha ao fechar o arquivo de log", "error", err)
		}
	}()
	slog.SetDefault(log)

	db, err := postgres.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() {
		if err := postgres.Close(db); err != nil {
			log.Error("falha ao fechar conexão", "error", err)
		}
	}()

	partRepository := postgres.NewPartRepository(db)

	partService := service.NewPartService(partRepository)
	restockService := service.NewRestockService(partRepository)

	app := httpapi.NewRouter(httpapi.RouterConfig{
		Part:    httpapi.NewPartHandler(partService),
		Restock: httpapi.NewRestockHandler(restockService),
		Health:  httpapi.NewHealthHandler(postgres.Pinger(db)),
		Logger:  log,
	})

	serverErr := make(chan error, 1)
	go func() {
		log.Info("servidor iniciado", "port", cfg.Port, "log_level", cfg.Log.Level)
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
		log.Info("encerrando servidor")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return app.ShutdownWithContext(ctx)
}

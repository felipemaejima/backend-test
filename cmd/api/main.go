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

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	db, err := postgres.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := postgres.Close(db); err != nil {
			slog.Error("falha ao fechar conexão", "error", err)
		}
	}()

	partRepository := postgres.NewPartRepository(db)

	partService := service.NewPartService(partRepository)
	restockService := service.NewRestockService(partRepository)

	partHandler := httpapi.NewPartHandler(partService)
	restockHandler := httpapi.NewRestockHandler(restockService)

	app := httpapi.NewRouter(partHandler, restockHandler)

	go func() {
		slog.Info("servidor iniciado", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("servidor encerrou com erro", "error", err)
			os.Exit(1)
		}
	}()

	// Encerramento gracioso: drena as requisições em voo antes de sair.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("encerrando servidor")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("erro no shutdown", "error", err)
	}
}

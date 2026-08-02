package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

const healthPath = "/health"

type RouterConfig struct {
	Part    *PartHandler
	Restock *RestockHandler
	Health  *HealthHandler
	Logger  *slog.Logger
}

func NewRouter(cfg RouterConfig) *fiber.App {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}

	app := fiber.New(fiber.Config{
		AppName:               "restock-api",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		ErrorHandler:          errorHandler,
	})

	app.Use(
		requestid.New(requestid.Config{Generator: uuid.NewString}),
		requestLogger(log),
		recovermw.New(),
	)

	app.Get(healthPath, cfg.Health.Health)

	parts := app.Group("/parts")
	parts.Post("/", cfg.Part.Create)
	parts.Get("/", cfg.Part.List)
	parts.Get("/:id", cfg.Part.GetByID)
	parts.Put("/:id", cfg.Part.Update)
	parts.Delete("/:id", cfg.Part.Delete)

	restock := app.Group("/restock")
	restock.Get("/priorities", cfg.Restock.Priorities)

	return app
}

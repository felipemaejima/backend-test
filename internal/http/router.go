
package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func NewRouter(partHandler *PartHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "restock-api",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		ErrorHandler:          errorHandler,
	})

	app.Use(requestid.New(), recovermw.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	parts := app.Group("/parts")
	parts.Post("/", partHandler.Create)
	parts.Get("/", partHandler.List)
	parts.Get("/:id", partHandler.GetByID)
	parts.Put("/:id", partHandler.Update)
	parts.Delete("/:id", partHandler.Delete)

	return app
}

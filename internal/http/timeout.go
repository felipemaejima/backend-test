package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func requestTimeout(limit time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), limit)
		defer cancel()

		c.SetUserContext(ctx)
		return c.Next()
	}
}

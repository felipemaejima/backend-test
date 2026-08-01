package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

const healthCheckTimeout = 2 * time.Second

type HealthChecker func(ctx context.Context) error

type HealthHandler struct {
	check HealthChecker
}

func NewHealthHandler(check HealthChecker) *HealthHandler {
	return &HealthHandler{check: check}
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), healthCheckTimeout)
	defer cancel()

	if err := h.check(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unavailable",
			"database": "down",
		})
	}

	return c.JSON(fiber.Map{"status": "ok", "database": "up"})
}

package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func requestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		chainErr := c.Next()
		if chainErr != nil {
			if err := errorHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		status := c.Response().StatusCode()

		attrs := []any{
			slog.String("request_id", requestID(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.String("ip", c.IP()),
		}
		if chainErr != nil {
			attrs = append(attrs, slog.String("error", chainErr.Error()))
		}

		log.Log(c.Context(), levelFor(c.Path(), status), "requisição", attrs...)

		return nil
	}
}

func levelFor(path string, status int) slog.Level {
	switch {
	case status >= fiber.StatusInternalServerError:
		return slog.LevelError
	case status >= fiber.StatusBadRequest:
		return slog.LevelWarn
	case path == healthPath:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

func requestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(requestid.ConfigDefault.ContextKey).(string); ok {
		return id
	}
	return ""
}

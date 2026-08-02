package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type errorResponse struct {
	Error  string              `json:"error"`
	Fields []domain.FieldError `json:"fields,omitempty"`
}

func errorHandler(c *fiber.Ctx, err error) error {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse{
			Error:  "dados inválidos",
			Fields: validationErr.Fields,
		})
	}

	if errors.Is(err, domain.ErrPartNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: err.Error()})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(errorResponse{Error: fiberErr.Message})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
		Error: "erro interno",
	})
}

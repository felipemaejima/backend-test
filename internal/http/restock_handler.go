package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/felipemaejima/backend-test/internal/service"
)

type RestockHandler struct {
	service *service.RestockService
}

func NewRestockHandler(svc *service.RestockService) *RestockHandler {
	return &RestockHandler{service: svc}
}

func (h *RestockHandler) Priorities(c *fiber.Ctx) error {
	page, err := h.service.Priorities(c.UserContext(), pageRequest(c))
	if err != nil {
		return err
	}

	return c.JSON(newRestockPrioritiesResponse(page))
}

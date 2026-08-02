package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/service"
)

type PartHandler struct {
	service *service.PartService
}

func NewPartHandler(svc *service.PartService) *PartHandler {
	return &PartHandler{service: svc}
}

// Create trata POST /parts.
func (h *PartHandler) Create(c *fiber.Ctx) error {
	var req partRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "corpo da requisição inválido")
	}

	part, err := h.service.Create(c.Context(), req.toInput())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(newPartResponse(*part))
}

// List trata GET /parts?category=&limit=&offset=.
func (h *PartHandler) List(c *fiber.Ctx) error {
	filter := domain.PartFilter{
		Category: c.Query("category"),
		Page:     pageRequest(c),
	}

	page, err := h.service.List(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(newPartListResponse(page))
}

// GetByID trata GET /parts/:id.
func (h *PartHandler) GetByID(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	part, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(newPartResponse(*part))
}

// Update trata PUT /parts/:id.
func (h *PartHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req partRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "corpo da requisição inválido")
	}

	part, err := h.service.Update(c.Context(), id, req.toInput())
	if err != nil {
		return err
	}

	return c.JSON(newPartResponse(*part))
}

// Delete trata DELETE /parts/:id.
func (h *PartHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func parseID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	return id, nil
}

func queryInt(c *fiber.Ctx, key string) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return 0
	}
	return value
}

func pageRequest(c *fiber.Ctx) domain.PageRequest {
	return domain.PageRequest{
		Number: queryInt(c, "page"),
		Size:   queryInt(c, "n"),
	}
}

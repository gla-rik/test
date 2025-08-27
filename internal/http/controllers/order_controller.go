package controllers

import (
	"wb/internal/config/database/postgre"

	"github.com/gofiber/fiber/v2"
)

type OrderController struct {
	db *postgre.Database
}

// NewOrderController создает новый контроллер заказов
func NewOrderController(db *postgre.Database) *OrderController {
	return &OrderController{
		db: db,
	}
}

// Ping живой?
func (oc *OrderController) Ping(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

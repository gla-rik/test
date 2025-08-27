package routes

import (
	"wb/config"
	"wb/internal/config/database/postgre"
	"wb/internal/http/controllers"

	"github.com/gofiber/fiber/v2"
)

type Router struct {
	app             *fiber.App
	cfg             *config.Config
	db              *postgre.Database
	orderController *controllers.OrderController
}

func NewRouter(app *fiber.App, cfg *config.Config, db *postgre.Database, orderController *controllers.OrderController) *Router {
	router := &Router{
		app:             app,
		cfg:             cfg,
		db:              db,
		orderController: orderController,
	}

	router.setupRoutes()

	return router
}

func (r *Router) setupRoutes() {
	r.app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// API маршруты
	_ = r.app.Group("/api")
}

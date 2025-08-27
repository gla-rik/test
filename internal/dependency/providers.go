package dependency

import (
	"errors"

	"wb/config"
	"wb/internal/config/database/postgre"
	"wb/internal/http/controllers"
	"wb/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet( //nolint:gochecknoglobals
	config.LoadConfig,

	postgre.NewDatabase,

	controllers.NewOrderController,

	routes.NewRouter,

	NewFiberApp,

	wire.Struct(new(App), "*"),
)

type App struct {
	FiberApp *fiber.App
	Router   *routes.Router
	Config   *config.Config
}

func NewApp(fiberApp *fiber.App, router *routes.Router, cfg *config.Config) *App {
	return &App{
		FiberApp: fiberApp,
		Router:   router,
		Config:   cfg,
	}
}

func NewFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      "WB Test App",
		ServerHeader: "WB Test Server",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			e := &fiber.Error{}
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})
}

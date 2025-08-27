package dependency

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"wb/config"
	"wb/internal/config/database/postgre"
	"wb/internal/http/controllers"
	httpErrors "wb/internal/http/errors"
	"wb/internal/http/middleware"
	"wb/internal/orm/repositories"
	"wb/internal/routes"
	"wb/internal/services"
)

// Провайдер для извлечения KafkaConfig из Config
func ProvideKafkaConfig(cfg *config.Config) *config.KafkaConfig {
	return cfg.Kafka
}

var ProviderSet = wire.NewSet( //nolint:gochecknoglobals
	config.LoadConfig,

	postgre.NewDatabase,

	// Провайдеры для конфигурации
	ProvideKafkaConfig,

	// Репозитории
	repositories.NewOrderRepository,

	// Сервисы
	services.NewCacheService,
	services.NewKafkaService,
	services.NewFakeDataService,

	// Контроллеры
	controllers.NewOrderController,
	controllers.NewKafkaController,

	// Роутеры
	routes.NewRouter,

	NewFiberApp,

	wire.Struct(new(App), "*"),
)

type App struct {
	FiberApp *fiber.App
	Router   *routes.Router
	Config   *config.Config
	Kafka    *services.KafkaService
	Cache    *services.CacheService
	FakeData *services.FakeDataService
}

func NewApp(fiberApp *fiber.App, router *routes.Router, cfg *config.Config, kafka *services.KafkaService, cache *services.CacheService, fakeData *services.FakeDataService) *App {
	return &App{
		FiberApp: fiberApp,
		Router:   router,
		Config:   cfg,
		Kafka:    kafka,
		Cache:    cache,
		FakeData: fakeData,
	}
}

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "WB Test App",
		ServerHeader: "WB Test Server",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			e := &fiber.Error{}
			if errors.As(err, &e) {
				code = e.Code
			}

			// centralized mapping
			mappedCode, message := httpErrors.MapErrorToStatus(err)
			if mappedCode != 0 {
				code = mappedCode
			}

			return ctx.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": message,
			})
		},
	})

	// Добавляем recovery middleware для обработки паник
	app.Use(middleware.Recovery())

	return app
}

package routes

import (
	"github.com/gofiber/fiber/v2"
	"wb/internal/http/controllers"
)

type Router struct {
	app             *fiber.App
	orderController *controllers.Order
	kafkaController *controllers.KafkaController
}

func NewRouter(app *fiber.App, orderController *controllers.Order, kafkaController *controllers.KafkaController) *Router {
	router := &Router{
		app:             app,
		orderController: orderController,
		kafkaController: kafkaController,
	}

	router.setupRoutes()

	return router
}

func (r *Router) setupRoutes() {
	if r.app == nil {
		return
	}

	// Базовый маршрут для проверки здоровья
	r.app.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})

	// Корневой маршрут для главной страницы
	r.app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendFile("./static/index.html")
	})

	// Раздача статических файлов
	r.app.Static("/", "./static")

	// API маршруты
	api := r.app.Group("/api")

	// Маршруты для заказов
	orders := api.Group("/orders")
	orders.Get("/", r.orderController.ListOrders)                  // GET /api/orders
	orders.Get("/uid/:uid", r.orderController.GetOrderByUIDFromDB) // GET /api/orders/uid/abc123

	// Маршруты для кеша
	cache := api.Group("/cache")
	cache.Get("/stats", func(ctx *fiber.Ctx) error {
		// Получаем статистику кеша через контроллер
		return r.orderController.GetCacheStats(ctx)
	})

	// Маршруты для фейковых данных
	fake := api.Group("/fake")
	fake.Post("/generate", func(ctx *fiber.Ctx) error {
		// Генерируем и отправляем фейковый заказ
		return r.orderController.GenerateFakeOrder(ctx)
	})

	// Маршруты для Kafka
	kafka := api.Group("/kafka")
	kafka.Post("/start", r.kafkaController.StartKafkaConsumer)      // POST /api/kafka/start
	kafka.Post("/stop", r.kafkaController.StopKafkaConsumer)        // POST /api/kafka/stop
	kafka.Get("/status", r.kafkaController.GetKafkaStatus)          // GET /api/kafka/status
	kafka.Post("/send", r.kafkaController.SendTestMessage)          // POST /api/kafka/send
	kafka.Post("/handler", r.kafkaController.RegisterCustomHandler) // POST /api/kafka/handler
}

// SetupRoutes настраивает маршруты для переданного приложения
func (r *Router) SetupRoutes(app *fiber.App) {
	r.app = app
	r.setupRoutes()
}

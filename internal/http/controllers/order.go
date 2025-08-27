package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"wb/internal/orm/repositories"
	"wb/internal/services"
)

type Order struct {
	db        *gorm.DB
	cache     *services.CacheService
	orderRepo *repositories.OrderRepository
}

// NewOrderController создает новый контроллер заказов
func NewOrderController(
	db *gorm.DB,
	cache *services.CacheService,
	orderRepo *repositories.OrderRepository,
) *Order {
	return &Order{
		db:        db,
		cache:     cache,
		orderRepo: orderRepo,
	}
}

func (oc *Order) Ping(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

// ListOrders возвращает список всех заказов (заглушка)
func (oc *Order) ListOrders(ctx *fiber.Ctx) error {
	orders := oc.cache.GetAllOrders()
	if orders != nil {
		log.Println("Данные с кэша")

		return ctx.Status(fiber.StatusOK).JSON(orders)
	}

	log.Println("В кэше данных нет, проверяем в базе")

	orders, err := oc.orderRepo.ListAll()
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(orders)
}

// GetOrderByUIDFromDB получает заказ по UID из базы данных
func (oc *Order) GetOrderByUIDFromDB(ctx *fiber.Ctx) error {
	orderUID := ctx.Params("uid")
	if orderUID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "UID заказа обязателен",
		})
	}

	order, ok := oc.cache.GetOrder(orderUID)
	if ok {
		log.Println("Данные с кэша")

		return ctx.Status(fiber.StatusOK).JSON(order)
	}

	log.Println("В кэше данных нет, проверяем в базе")

	order, err := oc.orderRepo.GetOrderByUID(orderUID)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(order)
}

// GetCacheStats возвращает статистику кеша
func (oc *Order) GetCacheStats(ctx *fiber.Ctx) error {
	stats := oc.cache.GetCacheStats()
	return ctx.Status(fiber.StatusOK).JSON(stats)
}

// GenerateFakeOrder генерирует и отправляет фейковый заказ
func (oc *Order) GenerateFakeOrder(ctx *fiber.Ctx) error {
	// Создаем временный FakeDataService для генерации
	fakeService := services.NewFakeDataService()

	// Генерируем фейковый заказ
	order := fakeService.GenerateFakeOrder()

	// Сохраняем в БД и кеш
	if err := oc.cache.SaveOrderToDB(order); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при сохранении заказа: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Фейковый заказ успешно создан",
		"order_uid": order.OrderUID,
		"order":     order,
	})
}

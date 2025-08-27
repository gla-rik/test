package controllers

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"wb/internal/services"
)

type KafkaController struct {
	kafkaService *services.KafkaService
}

func NewKafkaController(kafkaService *services.KafkaService) *KafkaController {
	return &KafkaController{
		kafkaService: kafkaService,
	}
}

// StartKafkaConsumer запускает потребителя Kafka
func (kc *KafkaController) StartKafkaConsumer(ctx *fiber.Ctx) error {
	if err := kc.kafkaService.Connect(); err != nil {
		log.Printf("Ошибка подключения к Kafka: %v", err)

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Ошибка подключения к Kafka: " + err.Error(),
		})
	}

	if err := kc.kafkaService.StartConsuming(); err != nil {
		log.Printf("Ошибка запуска потребителя Kafka: %v", err)

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Ошибка запуска потребителя Kafka: " + err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"error":   false,
		"message": "Kafka потребитель успешно запущен",
		"status":  "running",
	})
}

// StopKafkaConsumer останавливает потребителя Kafka
func (kc *KafkaController) StopKafkaConsumer(ctx *fiber.Ctx) error {
	if err := kc.kafkaService.Stop(); err != nil {
		log.Printf("Ошибка остановки Kafka: %v", err)

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Ошибка остановки Kafka: " + err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"error":   false,
		"message": "Kafka потребитель успешно остановлен",
		"status":  "stopped",
	})
}

// GetKafkaStatus возвращает статус Kafka сервиса
func (kc *KafkaController) GetKafkaStatus(ctx *fiber.Ctx) error {
	status := kc.kafkaService.GetStatus()

	return ctx.JSON(fiber.Map{
		"error":  false,
		"status": status,
	})
}

// SendTestMessage отправляет тестовое сообщение в Kafka
func (kc *KafkaController) SendTestMessage(ctx *fiber.Ctx) error {
	var request struct {
		Topic   string          `json:"topic"`
		Key     string          `json:"key"`
		Message json.RawMessage `json:"message"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Неверный формат запроса: " + err.Error(),
		})
	}

	if request.Topic == "" {
		request.Topic = "orders"
	}

	if request.Key == "" {
		request.Key = "test-message"
	}

	if request.Message == nil {
		// Создаем тестовое сообщение по умолчанию
		testMessage := map[string]interface{}{
			"order_id":   "test-order-123",
			"user_id":    "test-user-456",
			"status":     "created",
			"created_at": "2024-01-01T00:00:00Z",
			"message":    "Это тестовое сообщение",
		}

		messageBytes, err := json.Marshal(testMessage)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Ошибка создания тестового сообщения: " + err.Error(),
			})
		}

		request.Message = messageBytes
	}

	if err := kc.kafkaService.SendMessage(request.Topic, request.Key, request.Message); err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Ошибка отправки сообщения: " + err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"error":   false,
		"message": "Тестовое сообщение успешно отправлено",
		"topic":   request.Topic,
		"key":     request.Key,
	})
}

// RegisterCustomHandler регистрирует пользовательский обработчик для топика
func (kc *KafkaController) RegisterCustomHandler(ctx *fiber.Ctx) error {
	var request struct {
		Topic   string `json:"topic"`
		Handler string `json:"handler_type"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Неверный формат запроса: " + err.Error(),
		})
	}

	if request.Topic == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Топик не указан",
		})
	}

	// Регистрируем обработчик в зависимости от типа
	switch request.Handler {
	case "log":
		kc.kafkaService.RegisterHandler(request.Topic, func(message []byte) error {
			log.Printf("Пользовательский обработчик для топика %s: %s", request.Topic, string(message))
			return nil
		})
	case "json":
		kc.kafkaService.RegisterHandler(request.Topic, func(message []byte) error {
			var data interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Printf("Ошибка парсинга JSON из топика %s: %v", request.Topic, err)
				return err
			}

			log.Printf("JSON сообщение из топика %s: %+v", request.Topic, data)

			return nil
		})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Неизвестный тип обработчика. Поддерживаемые типы: log, json",
		})
	}

	return ctx.JSON(fiber.Map{
		"error":   false,
		"message": "Пользовательский обработчик успешно зарегистрирован",
		"topic":   request.Topic,
		"handler": request.Handler,
	})
}

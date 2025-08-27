package middleware

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RecoveryConfig содержит настройки для recovery middleware
type RecoveryConfig struct {
	EnableStackTrace bool
	LogPanic         bool
	CustomHandler    func(ctx *fiber.Ctx, panicVal interface{})
}

// Recovery создает recovery middleware с настройками по умолчанию
func Recovery() fiber.Handler {
	return RecoveryWithConfig(RecoveryConfig{
		EnableStackTrace: true,
		LogPanic:         true,
	})
}

// RecoveryWithConfig создает recovery middleware с кастомными настройками
func RecoveryWithConfig(config RecoveryConfig) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				if config.LogPanic {
					log.Printf("[PANIC RECOVERED] %s %s - %v", ctx.Method(), ctx.Path(), r)

					if config.EnableStackTrace {
						log.Printf("[STACK TRACE] %s", debug.Stack())
					}
				}

				if config.CustomHandler != nil {
					config.CustomHandler(ctx, r)
					return
				}

				handlePanic(ctx, r, config)
			}
		}()

		return ctx.Next()
	}
}

// handlePanic обрабатывает панику стандартным способом
func handlePanic(ctx *fiber.Ctx, panicVal interface{}, config RecoveryConfig) {
	// Создаем детальный ответ об ошибке
	errorResponse := fiber.Map{
		"error":     true,
		"message":   "Внутренняя ошибка сервера",
		"timestamp": time.Now().Format(time.RFC3339),
		"path":      ctx.Path(),
		"method":    ctx.Method(),
		"panic":     fmt.Sprintf("%v", panicVal),
	}

	// Добавляем stack trace если включен
	if config.EnableStackTrace {
		errorResponse["stack_trace"] = string(debug.Stack())
	}

	// Отправляем ответ
	if err := ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse); err != nil {
		// Если не удалось отправить JSON, отправляем простой текст
		_ = ctx.Status(fiber.StatusInternalServerError).SendString("Внутренняя ошибка сервера")
	}
}

// RecoveryWithLogger создает recovery middleware с кастомным логгером
func RecoveryWithLogger(logger *log.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("[PANIC RECOVERED] %s %s - %v", ctx.Method(), ctx.Path(), r)
				logger.Printf("[STACK TRACE] %s", debug.Stack())

				handlePanic(ctx, r, RecoveryConfig{EnableStackTrace: true, LogPanic: true})
			}
		}()

		return ctx.Next()
	}
}

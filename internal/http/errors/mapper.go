package errors

import (
	goerrors "errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// MapErrorToStatus возвращает соответствующий код статуса HTTP и безопасное для клиента сообщение для данной ошибки.
// Он сохраняет коды fiber.Error, отображает общие состояния базы данных и возвращается к коду 500.
func MapErrorToStatus(err error) (int, string) {
	if err == nil {
		return fiber.StatusOK, ""
	}

	// если error это -  *fiber.Error, то - используем её
	var fe *fiber.Error
	if goerrors.As(err, &fe) {
		return fe.Code, fe.Message
	}

	if goerrors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.StatusNotFound, "Ресурс не найден"
	}

	return fiber.StatusInternalServerError, err.Error()
}

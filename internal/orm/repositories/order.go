package repositories

import (
	"log"

	"wb/internal/orm/models"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// OrderRepository репозиторий для работы с заказами
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository создает новый экземпляр репозитория
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrderByUID(orderUID string) (
	*models.Order,
	error,
) {
	order := models.Order{}

	err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Where("order_uid = ?", orderUID).
		First(&order).Error
	if err != nil {
		return nil, eris.Wrap(err, err.Error())
	}

	return &order, nil
}

// ListAll возвращает список всех заказов (заглушка)
func (r *OrderRepository) ListAll() ([]models.Order, error) {
	var orders []models.Order
	if err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error; err != nil {
		return nil, eris.Wrap(err, "ошибка при получении всех заказов")
	}

	return orders, nil
}

// CreateWithRelations создает заказ со всеми связанными данными
func (r *OrderRepository) CreateWithRelations(order *models.Order) error {
	log.Printf("Начинаем создание заказа с UID: %s", order.OrderUID)

	tx := r.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Паника при создании заказа: %v", r)
			tx.Rollback()
		}
	}()

	// Сбрасываем ID для основного заказа, чтобы использовать автоинкремент
	order.ID = 0
	log.Printf("Сброшен ID основного заказа, теперь ID = %d", order.ID)

	// Создаем основной заказ
	if err := tx.Create(order).Error; err != nil {
		log.Printf("Ошибка создания основного заказа: %v", err)
		tx.Rollback()

		return eris.Wrap(err, err.Error())
	}

	log.Printf("Основной заказ создан с ID: %d", order.ID)

	if order.Delivery != nil {
		// Сбрасываем ID для доставки, чтобы использовать автоинкремент
		order.Delivery.ID = 0
		order.Delivery.OrderID = order.ID
		log.Printf("Создаем доставку для заказа ID: %d", order.ID)

		if err := tx.Create(order.Delivery).Error; err != nil {
			log.Printf("Ошибка создания доставки: %v", err)
			tx.Rollback()

			return eris.Wrap(err, err.Error())
		}

		log.Printf("Доставка создана с ID: %d", order.Delivery.ID)
	}

	if order.Payment != nil {
		// Сбрасываем ID для платежа, чтобы использовать автоинкремент
		order.Payment.ID = 0
		order.Payment.OrderID = order.ID
		log.Printf("Создаем платеж для заказа ID: %d", order.ID)

		if err := tx.Create(order.Payment).Error; err != nil {
			log.Printf("Ошибка создания платежа: %v", err)
			tx.Rollback()

			return eris.Wrap(err, err.Error())
		}

		log.Printf("Платеж создан с ID: %d", order.Payment.ID)
	}

	if len(order.Items) > 0 {
		log.Printf("Создаем %d товаров для заказа ID: %d", len(order.Items), order.ID)

		for i := range order.Items {
			// Сбрасываем ID для каждого товара, чтобы использовать автоинкремент
			order.Items[i].ID = 0

			order.Items[i].OrderID = order.ID
			if err := tx.Create(&order.Items[i]).Error; err != nil {
				log.Printf("Ошибка создания товара %d: %v", i+1, err)
				tx.Rollback()

				return eris.Wrap(err, err.Error())
			}

			log.Printf("Товар %d создан с ID: %d", i+1, order.Items[i].ID)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Ошибка коммита транзакции: %v", err)
		return eris.Wrap(err, err.Error())
	}

	log.Printf("Заказ успешно создан и сохранен в базе данных")

	return nil
}

// ClearAll очищает все данные
func (r *OrderRepository) ClearAll() error {
	// Используем TRUNCATE для полной очистки таблиц и сброса последовательностей
	// Это более эффективно чем DELETE и автоматически сбрасывает последовательности
	if err := r.db.Exec("TRUNCATE TABLE order_items, payments, deliveries, orders RESTART IDENTITY CASCADE").Error; err != nil {
		return eris.Wrap(err, err.Error())
	}

	return nil
}

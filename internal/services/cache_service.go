package services

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"wb/internal/orm/models"
	"wb/internal/orm/repositories"
)

type CacheService struct {
	orders map[string]*models.Order
	mu     sync.RWMutex
	db     *gorm.DB
}

func NewCacheService(db *gorm.DB) *CacheService {
	service := &CacheService{
		orders: make(map[string]*models.Order),
		db:     db,
	}

	// Восстанавливаем кеш из БД при старте
	service.RestoreFromDB()

	return service
}

// SetOrder сохраняет заказ в кеш
func (cs *CacheService) SetOrder(order *models.Order) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Используем OrderUID как ключ для кеша
	if order.OrderUID != "" {
		cs.orders[order.OrderUID] = order
		log.Printf("Заказ %s добавлен в кеш", order.OrderUID)
	}
}

// GetOrder получает заказ из кеша по OrderUID
func (cs *CacheService) GetOrder(orderUID string) (*models.Order, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	order, exists := cs.orders[orderUID]

	return order, exists
}

// GetOrderByID получает заказ из кеша по ID
func (cs *CacheService) GetOrderByID(orderID string) (*models.Order, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for _, order := range cs.orders {
		if order.OrderUID == orderID {
			return order, true
		}
	}

	return nil, false
}

// GetAllOrders возвращает все заказы из кеша
func (cs *CacheService) GetAllOrders() []models.Order {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	orders := make([]models.Order, 0, len(cs.orders))
	for _, order := range cs.orders {
		orders = append(orders, *order)
	}

	if len(orders) == 0 {
		return nil
	}

	return orders
}

// RestoreFromDB восстанавливает кеш из базы данных
func (cs *CacheService) RestoreFromDB() {
	log.Println("Восстановление кеша из базы данных...")

	var orders []models.Order

	err := cs.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error
	if err != nil {
		log.Printf("Ошибка при восстановлении кеша из БД: %v", err)
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Очищаем текущий кеш
	cs.orders = make(map[string]*models.Order)

	// Заполняем кеш данными из БД
	for i := range orders {
		if orders[i].OrderUID != "" {
			cs.orders[orders[i].OrderUID] = &orders[i]
		}
	}

	log.Printf("Кеш восстановлен из БД: %d заказов", len(cs.orders))
}

// SaveOrderToDB сохраняет заказ в базу данных и обновляет кеш
func (cs *CacheService) SaveOrderToDB(order *models.Order) error {
	// Сохраняем в БД cо всеми связями через репозиторий
	repo := repositories.NewOrderRepository(cs.db)

	err := repo.CreateWithRelations(order)
	if err != nil {
		log.Printf("Ошибка при сохранении заказа и связей в БД: %v", err)
		return err
	}

	// Обновляем кеш
	cs.SetOrder(order)

	log.Printf("Заказ %s сохранен в БД и добавлен в кеш", order.OrderUID)

	return nil
}

// GetCacheStats возвращает статистику кеша
func (cs *CacheService) GetCacheStats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return map[string]interface{}{
		"total_orders": len(cs.orders),
		"cache_size":   len(cs.orders),
		"last_updated": time.Now().Format(time.RFC3339),
	}
}

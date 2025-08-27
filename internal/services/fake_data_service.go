//nolint:mnd
package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/rotisserie/eris"
	"wb/internal/orm/models"
)

type FakeDataService struct {
	kafkaService *KafkaService
	orderCounter int
}

func NewFakeDataService() *FakeDataService {
	// Инициализируем генератор случайных чисел
	gofakeit.Seed(time.Now().UnixNano())

	return &FakeDataService{
		orderCounter: 1,
	}
}

// SetKafkaService устанавливает KafkaService после создания
func (fds *FakeDataService) SetKafkaService(kafkaService *KafkaService) {
	fds.kafkaService = kafkaService
}

// GenerateFakeOrder генерирует фейковый заказ
func (fds *FakeDataService) GenerateFakeOrder() *models.Order {
	// Генерируем уникальный OrderUID
	orderUID := fmt.Sprintf("order_%d_%s", fds.orderCounter, gofakeit.UUID())
	fds.orderCounter++

	// Генерируем случайное количество товаров (1-5)
	itemCount := gofakeit.Number(1, 5)

	var (
		items       []models.OrderItem
		totalAmount float64
	)

	for i := 0; i < itemCount; i++ {
		price := float64(gofakeit.IntRange(100, 5000))
		sale := gofakeit.Number(0, 50) // Скидка 0-50%
		totalPrice := price * (1 - float64(sale)/100)

		item := models.OrderItem{
			ChrtID:      gofakeit.IntRange(1000000, 9999999),
			TrackNumber: orderUID,
			Price:       price,
			Rid:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        sale,
			Size:        gofakeit.RandomString([]string{"XS", "S", "M", "L", "XL", "XXL", "0", "1", "2", "3"}),
			TotalPrice:  totalPrice,
			NmID:        gofakeit.IntRange(1000000, 9999999),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.IntRange(100, 999),
		}
		items = append(items, item)
		totalAmount += totalPrice
	}

	// Генерируем информацию о доставке
	delivery := &models.Delivery{
		Name:    gofakeit.Name(),
		Phone:   gofakeit.Phone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Street() + " " + gofakeit.StreetNumber(),
		Region:  gofakeit.State(),
		Email:   gofakeit.Email(),
	}

	// Генерируем информацию о платеже
	payment := &models.Payment{
		Transaction:  gofakeit.UUID(),
		RequestID:    gofakeit.UUID(),
		Currency:     gofakeit.RandomString([]string{"USD", "EUR", "RUB"}),
		Provider:     gofakeit.RandomString([]string{"wbpay", "stripe", "paypal", "yandex"}),
		Amount:       totalAmount + 1500, // Общая сумма + стоимость доставки
		PaymentDt:    gofakeit.Date(),
		Bank:         gofakeit.RandomString([]string{"Альфа-Банк", "Сбер", "Тиньк", "ВТБ", "ФПИ Банк"}),
		DeliveryCost: 1500,
		GoodsTotal:   totalAmount,
		CustomFee:    0,
	}

	// Создаем заказ
	order := &models.Order{
		OrderUID:          orderUID,
		TrackNumber:       fmt.Sprintf("WB%s", gofakeit.RandomString([]string{"ILM", "TEST", "TRACK"})),
		Entry:             "WBIL",
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            gofakeit.RandomString([]string{"Ростов", "Не Ростов"}),
		InternalSignature: "",
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.RandomString([]string{"сдек", "почта россии", "dhl", "голубь"}),
		ShardKey:          fmt.Sprintf("%d", gofakeit.IntRange(1, 10)),
		SmID:              gofakeit.IntRange(1, 100),
		DateCreated:       time.Now(),
		OofShard:          fmt.Sprintf("%d", gofakeit.IntRange(1, 5)),
		TotalAmount:       totalAmount,
	}

	return order
}

// SendFakeOrderToKafka отправляет фейковый заказ в Kafka
func (fds *FakeDataService) SendFakeOrderToKafka() error {
	order := fds.GenerateFakeOrder()

	// Конвертируем в JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге заказа: %w", err)
	}

	// Отправляем в Kafka
	err = fds.kafkaService.SendMessage("orders", order.OrderUID, orderJSON)
	if err != nil {
		return fmt.Errorf("ошибка при отправке в Kafka: %w", err)
	}

	log.Printf("Фейковый заказ %s отправлен в Kafka", order.OrderUID)

	return nil
}

// StartFakeDataGeneration запускает периодическую генерацию фейковых данных
func (fds *FakeDataService) StartFakeDataGeneration(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Запущена генерация фейковых данных каждые %v", interval)

	<-ticker.C

	if err := fds.SendFakeOrderToKafka(); err != nil {
		log.Printf("Ошибка при отправке фейкового заказа: %v", err)
	}
}

// LoadTestDataFromFile загружает тестовые данные из JSON файла
func (fds *FakeDataService) LoadTestDataFromFile() error {
	// Создаем тестовый заказ на основе примера
	testOrder := &models.Order{
		ID:          1,
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: &models.Delivery{
			OrderID: 1,
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: &models.Payment{
			OrderID:      1,
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Unix(1637907727, 0),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.OrderItem{
			{
				OrderID:     1,
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
		TotalAmount:       317,
	}

	// Создаем сообщение в формате OrderMessage
	orderMessage := &OrderMessage{
		OrderID:   testOrder.OrderUID,
		UserID:    testOrder.CustomerID,
		Status:    "created",
		Items:     testOrder.Items,
		Payment:   *testOrder.Payment,
		Delivery:  *testOrder.Delivery,
		CreatedAt: testOrder.DateCreated,
		UpdatedAt: time.Now(),
	}

	// Отправляем тестовый заказ в Kafka
	orderJSON, err := json.Marshal(orderMessage)
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге тестового сообщения заказа: %w", err)
	}

	err = fds.kafkaService.SendMessage("orders", testOrder.OrderUID, orderJSON)
	if err != nil {
		return eris.Wrap(err, err.Error())
	}

	return nil
}

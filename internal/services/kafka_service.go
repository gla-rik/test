package services

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/rotisserie/eris"
	"wb/config"
	"wb/internal/orm/models"
)

const delayToRepeat = 5

type KafkaService struct {
	config    *config.KafkaConfig
	consumer  sarama.ConsumerGroup
	producer  sarama.SyncProducer
	isRunning bool
	mu        sync.RWMutex
	handlers  map[string]MessageHandler
	ctx       context.Context
	cancel    context.CancelFunc
	cache     *CacheService
}

type MessageHandler func(message []byte) error

type OrderMessage struct {
	OrderID   string                 `json:"order_id"`
	UserID    string                 `json:"user_id"`
	Status    string                 `json:"status"`
	Items     []models.OrderItem     `json:"items"`
	Payment   models.Payment         `json:"payment"`
	Delivery  models.Delivery        `json:"delivery"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewKafkaService(cfg *config.KafkaConfig, cache *CacheService) (*KafkaService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	service := &KafkaService{
		config:   cfg,
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
		cache:    cache,
	}

	// Инициализируем обработчики по умолчанию
	service.registerDefaultHandlers()

	return service, nil
}

func (k *KafkaService) registerDefaultHandlers() {
	// Обработчик для сообщений о заказах
	k.RegisterHandler("orders", func(message []byte) error {
		var orderMsg OrderMessage
		if err := json.Unmarshal(message, &orderMsg); err != nil {
			return eris.Wrapf(err, "failed to unmarshal order message")
		}

		// Базовая валидация
		if strings.TrimSpace(orderMsg.OrderID) == "" {
			return eris.New("order_id is required")
		}

		log.Printf("Получено сообщение о заказе: %s, статус: %s", orderMsg.OrderID, orderMsg.Status)

		// Преобразуем сообщение в модель Order и сохраняем в БД
		order := &models.Order{
			OrderUID:    orderMsg.OrderID,
			TrackNumber: orderMsg.OrderID, // Используем OrderID как TrackNumber
			CustomerID:  orderMsg.UserID,
			TotalAmount: 0, // Будет рассчитано из Items
		}

		// Копируем Items и считаем сумму
		var total int64

		for _, item := range orderMsg.Items {
			order.Items = append(order.Items, models.OrderItem{
				ChrtID:      item.ChrtID,
				TrackNumber: item.TrackNumber,
				Price:       item.Price,
				Rid:         item.Rid,
				Name:        item.Name,
				Sale:        item.Sale,
				Size:        item.Size,
				TotalPrice:  item.TotalPrice,
				NmID:        item.NmID,
				Brand:       item.Brand,
				Status:      item.Status,
			})

			if item.TotalPrice > 0 {
				total += int64(item.TotalPrice)
			} else {
				total += int64(item.Price)
			}
		}

		order.TotalAmount = float64(int(total))

		// Копируем Payment
		order.Payment = &models.Payment{
			Transaction:  orderMsg.Payment.Transaction,
			RequestID:    orderMsg.Payment.RequestID,
			Currency:     orderMsg.Payment.Currency,
			Provider:     orderMsg.Payment.Provider,
			Amount:       orderMsg.Payment.Amount,
			PaymentDt:    orderMsg.Payment.PaymentDt,
			Bank:         orderMsg.Payment.Bank,
			DeliveryCost: orderMsg.Payment.DeliveryCost,
			GoodsTotal:   orderMsg.Payment.GoodsTotal,
			CustomFee:    orderMsg.Payment.CustomFee,
		}

		// Копируем Delivery
		order.Delivery = &models.Delivery{
			Name:    orderMsg.Delivery.Name,
			Phone:   orderMsg.Delivery.Phone,
			Zip:     orderMsg.Delivery.Zip,
			City:    orderMsg.Delivery.City,
			Address: orderMsg.Delivery.Address,
			Region:  orderMsg.Delivery.Region,
			Email:   orderMsg.Delivery.Email,
		}

		// Сохраняем в БД и обновляем кеш
		if err := k.cache.SaveOrderToDB(order); err != nil {
			log.Printf("Ошибка при сохранении заказа: %v", err)
			return err
		}

		log.Printf("Заказ %s успешно обработан и сохранен", orderMsg.OrderID)

		return nil
	})
}

func (k *KafkaService) RegisterHandler(topic string, handler MessageHandler) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.handlers[topic] = handler
}

func (k *KafkaService) Connect() error {
	// Настройка конфигурации для consumer
	consumerConfig := sarama.NewConfig()
	// Версия брокера из конфигурации, по умолчанию 2.8.0
	consumerConfig.Version = sarama.V2_8_0_0
	if k.config.Version != "" {
		// простая карта поддерживаемых версий
		switch k.config.Version {
		case "2.8.0":
			consumerConfig.Version = sarama.V2_8_0_0
		case "3.2.0":
			consumerConfig.Version = sarama.V3_2_0_0
		}
	}

	consumerConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	if strings.ToLower(k.config.AutoOffset) == "latest" {
		consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerConfig.Consumer.Offsets.AutoCommit.Enable = true
	consumerConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	// Настройка конфигурации для producer
	producerConfig := sarama.NewConfig()
	producerConfig.Version = consumerConfig.Version
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = 3
	producerConfig.Producer.Return.Successes = true

	// Подключение к consumer group
	consumer, err := sarama.NewConsumerGroup(k.config.GetBrokers(), k.config.GetGroupID(), consumerConfig)
	if err != nil {
		return eris.Wrapf(err, "failed to create consumer group")
	}

	// Подключение к producer
	producer, err := sarama.NewSyncProducer(k.config.GetBrokers(), producerConfig)
	if err != nil {
		consumer.Close()
		return eris.Wrapf(err, "failed to create producer")
	}

	k.consumer = consumer
	k.producer = producer

	log.Printf("Успешно подключились к Kafka brokers: %s", k.config.GetBrokersString())

	return nil
}

func (k *KafkaService) StartConsuming() error {
	if k.consumer == nil {
		return eris.New("consumer не инициализирован, сначала вызовите Connect()")
	}

	k.mu.Lock()
	k.isRunning = true
	k.mu.Unlock()

	go func() {
		for {
			select {
			case <-k.ctx.Done():
				log.Println("Остановка потребления сообщений Kafka")
				return
			default:
				topics := []string{k.config.GetTopic()}

				err := k.consumer.Consume(k.ctx, topics, k)
				if err != nil {
					log.Printf("Ошибка при потреблении сообщений: %v", err)
					time.Sleep(delayToRepeat * time.Second) // Пауза перед повторной попыткой
				}
			}
		}
	}()

	log.Printf("Начато потребление сообщений из топика: %s", k.config.GetTopic())

	return nil
}

func (k *KafkaService) Stop() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.isRunning {
		return nil
	}

	k.isRunning = false
	k.cancel()

	if k.consumer != nil {
		if err := k.consumer.Close(); err != nil {
			log.Printf("Ошибка при закрытии consumer: %v", err)
		}
	}

	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			log.Printf("Ошибка при закрытии producer: %v", err)
		}
	}

	log.Println("Kafka сервис остановлен")

	return nil
}

func (k *KafkaService) SendMessage(topic string, key string, message []byte) error {
	if k.producer == nil {
		return eris.New("producer не инициализирован")
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return eris.Wrapf(err, "failed to send message to topic %s", topic)
	}

	log.Printf("Сообщение отправлено в топик %s, partition: %d, offset: %d", topic, partition, offset)

	return nil
}

// Реализация интерфейса sarama.ConsumerGroupHandler
func (k *KafkaService) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session setup")
	return nil
}

func (k *KafkaService) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session cleanup")
	return nil
}

func (k *KafkaService) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			log.Printf("Получено сообщение из топика %s, partition: %d, offset: %d",
				message.Topic, message.Partition, message.Offset)

			// Обработка сообщения
			if err := k.handleMessage(message.Topic, message.Value); err != nil {
				log.Printf("Ошибка обработки сообщения: %v", err)
				// В продакшене здесь можно добавить логику retry или dead letter queue
			}

			// Подтверждение обработки сообщения
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (k *KafkaService) handleMessage(topic string, message []byte) error {
	k.mu.RLock()
	handler, exists := k.handlers[topic]
	k.mu.RUnlock()

	if !exists {
		log.Printf("Обработчик для топика %s не найден", topic)
		return nil
	}

	return handler(message)
}

func (k *KafkaService) IsRunning() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.isRunning
}

func (k *KafkaService) GetStatus() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return map[string]interface{}{
		"is_running": k.isRunning,
		"brokers":    k.config.GetBrokers(),
		"topic":      k.config.GetTopic(),
		"group_id":   k.config.GetGroupID(),
		"connected":  k.consumer != nil && k.producer != nil,
	}
}

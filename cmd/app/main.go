package main

import (
	"fmt"
	"log"

	"wb/internal/dependency"
)

func main() {
	// Создаем приложение с помощью Wire
	app, err := dependency.InitializeApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	// Устанавливаем KafkaService в FakeDataService
	app.FakeData.SetKafkaService(app.Kafka)

	// Подключаемся к Kafka и запускаем consumer
	if err := app.Kafka.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Kafka: %v", err)
	} else {
		if err := app.Kafka.StartConsuming(); err != nil {
			log.Printf("Warning: Failed to start Kafka consumer: %v", err)
		} else {
			log.Println("Kafka consumer started successfully")
		}

		// Загружаем тестовые данные при старте
		if err := app.FakeData.LoadTestDataFromFile(); err != nil {
			log.Printf("Warning: Failed to load test data: %v", err)
		}
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", app.Config.App.Host, app.Config.App.Port)
	log.Printf("Starting server on %s", addr)

	if err := app.FiberApp.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

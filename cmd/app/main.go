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

	// Start server
	addr := fmt.Sprintf("%s:%s", app.Config.App.Host, app.Config.App.Port)
	log.Printf("Starting server on %s", addr)

	if err := app.FiberApp.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

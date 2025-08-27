package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rotisserie/eris"
)

type Config struct {
	App      *App
	Database *Database
	Kafka    *KafkaConfig
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, using defaults or environment variables: %v", err)
	}

	var cfg Config

	cfg.App = &App{}
	cfg.Database = &Database{}
	cfg.Kafka = &KafkaConfig{}

	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to process environment variables")
	}

	return &cfg, nil
}

package config

import (
	"strings"
	"time"
)

type KafkaConfig struct {
	Brokers        []string      `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	Topic          string        `envconfig:"KAFKA_TOPIC" default:"orders"`
	GroupID        string        `envconfig:"KAFKA_GROUP_ID" default:"wb-consumer-group"`
	Version        string        `envconfig:"KAFKA_VERSION" default:"2.8.0"`
	SessionTimeout time.Duration `envconfig:"KAFKA_SESSION_TIMEOUT" default:"30s"`
	AutoOffset     string        `envconfig:"KAFKA_AUTO_OFFSET" default:"earliest"`
	MaxWaitTime    time.Duration `envconfig:"KAFKA_MAX_WAIT_TIME" default:"1s"`
	MaxBytes       int           `envconfig:"KAFKA_MAX_BYTES" default:"1048576"`
}

func (k *KafkaConfig) GetBrokers() []string {
	if len(k.Brokers) == 0 {
		return []string{"localhost:9092"}
	}

	return k.Brokers
}

func (k *KafkaConfig) GetBrokersString() string {
	return strings.Join(k.GetBrokers(), ",")
}

func (k *KafkaConfig) GetTopic() string {
	if k.Topic == "" {
		return "orders"
	}

	return k.Topic
}

func (k *KafkaConfig) GetGroupID() string {
	if k.GroupID == "" {
		return "wb-consumer-group"
	}

	return k.GroupID
}

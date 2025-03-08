package avro_kafka_config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// KafkaClient permet de gérer les topics
type KafkaClient struct {
	config Config
}

// NewKafkaClient crée une instance du client Kafka
func NewKafkaClient(cfg Config) *KafkaClient {
	return &KafkaClient{config: cfg}
}

// CreateTopic crée un topic Kafka
func (kc *KafkaClient) CreateTopic(topic string, partitions int) error {
	dialer := &kafka.Dialer{
		SASLMechanism: plain.Mechanism{
			Username: kc.config.APIKey,
			Password: kc.config.APISecret,
		},
		TLS: &tls.Config{},
	}

	conn, err := dialer.DialContext(context.Background(), "tcp", kc.config.BootstrapServers)
	if err != nil {
		return fmt.Errorf("erreur de connexion à Kafka : %w", err)
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 3,
	})
	if err != nil {
		return fmt.Errorf("erreur lors de la création du topic %s : %w", topic, err)
	}

	log.Printf("Topic %s créé avec succès", topic)
	return nil
}

// ListTopics liste les topics disponibles
func (kc *KafkaClient) ListTopics() ([]string, error) {
	dialer := &kafka.Dialer{
		SASLMechanism: plain.Mechanism{
			Username: kc.config.APIKey,
			Password: kc.config.APISecret,
		},
		TLS: &tls.Config{},
	}

	conn, err := dialer.DialContext(context.Background(), "tcp", kc.config.BootstrapServers)
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à Kafka : %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture des partitions : %w", err)
	}

	topicsMap := make(map[string]bool)
	for _, p := range partitions {
		topicsMap[p.Topic] = true
	}

	var topics []string
	for topic := range topicsMap {
		topics = append(topics, topic)
	}

	return topics, nil
}

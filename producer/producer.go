package producer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/METAVENTUS/metaventus-kafka-adapters/models"
	"log"

	"github.com/hamba/avro"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Producer Kafka
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer initialise un Kafka Producer avec authentification
func NewProducer(ctx context.Context, cfg Config) (*Producer, error) {
	transport := &kafka.Transport{}

	if cfg.SASL {
		transport.SASL = plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	if cfg.TLS {
		transport.TLS = &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}
	}

	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("aucun broker spécifié")
	}

	primaryBroker := cfg.Brokers[0]

	conn, err := kafka.DialContext(ctx, "tcp", primaryBroker)
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion/ping sur %s: %w", primaryBroker, err)
	}
	defer conn.Close()

	w := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Topic:     cfg.Topic,
		Transport: transport,
		Balancer:  &kafka.LeastBytes{},
	}

	return &Producer{
		writer: w,
		topic:  cfg.Topic,
	}, nil
}

// Publish envoie un événement Avro au topic Kafka
func (p *Producer) Publish(ctx context.Context, event models.AvroEvent) error {
	schema := event.GetSchema()

	// Sérialisation Avro
	buf := new(bytes.Buffer)
	encoder, err := avro.NewEncoder(schema, buf)
	if err != nil {
		return fmt.Errorf("erreur encodeur Avro : %w", err)
	}

	if err = encoder.Encode(event); err != nil {
		return fmt.Errorf("erreur d'encodage Avro : %w", err)
	}

	// Envoi du message Kafka avec le topic spécifique
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.PartitionKey()),
		Value: buf.Bytes(),
	})
	if err != nil {
		return fmt.Errorf("erreur d'envoi Kafka : %w", err)
	}

	log.Printf("Message envoyé à %s avec clé %s\n", p.topic, event.PartitionKey())
	return nil
}

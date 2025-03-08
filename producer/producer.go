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
func NewProducer(cfg Config) *Producer {
	// Configuration de l'authentification SASL si activée
	var transport *kafka.Transport
	if cfg.SASL {
		transport = &kafka.Transport{
			SASL: plain.Mechanism{
				Username: cfg.Username,
				Password: cfg.Password,
			},
			TLS: &tls.Config{}, // Ajout de TLS si nécessaire
		}
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr:      kafka.TCP(cfg.Brokers...), // Connexion aux brokers
			Topic:     cfg.Topic,                 // Topic Kafka
			Transport: transport,                 // Ajout du transport avec SASL/TLS
		},
		topic: cfg.Topic,
	}
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

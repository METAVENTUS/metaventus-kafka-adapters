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
	// Créer un transport non-nil
	transport := &kafka.Transport{}

	if cfg.SASL {
		transport.SASL = plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	if cfg.TLS {
		transport.TLS = &tls.Config{}
	}

	// 1) "Ping" : vérifier la connectivité sur le premier broker
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("aucun broker spécifié")
	}
	primaryBroker := cfg.Brokers[0]

	// Dial direct pour checker si on arrive à se connecter
	conn, err := kafka.DialContext(ctx, "tcp", primaryBroker)
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion/ping sur %s: %w", primaryBroker, err)
	}
	defer conn.Close()

	// On peut pousser plus loin en faisant un read de métadonnées,
	// exemple: lire la liste des partitions du topic (si on veut être sûr que ce topic existe)
	_, err = conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("erreur ReadPartitions sur %s: %w", primaryBroker, err)
	}

	// 2) Construire le writer
	w := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Topic:     cfg.Topic,
		Transport: transport,
		// Autres options (BatchTimeout, Balancer, etc.) si besoin
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

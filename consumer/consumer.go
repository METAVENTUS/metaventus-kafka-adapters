package consumer

import (
	"bytes"
	"context"
	"crypto/tls"
	"log"
	"sync"

	"github.com/METAVENTUS/metaventus-kafka-adapters/models"
	"github.com/hamba/avro"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Consumer générique Kafka
type Consumer[T models.AvroEvent] struct {
	reader *kafka.Reader
}

// NewConsumer initialise un Kafka Consumer générique avec un type `T`
func NewConsumer[T models.AvroEvent](cfg Config) *Consumer[T] {
	// Configuration de l'authentification SASL si activée
	var dialer *kafka.Dialer
	if cfg.SASL {
		dialer = &kafka.Dialer{
			SASLMechanism: plain.Mechanism{
				Username: cfg.Username,
				Password: cfg.Password,
			},
			TLS: &tls.Config{},
		}
	}

	return &Consumer[T]{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  cfg.Brokers,
			Topic:    cfg.Topic,
			GroupID:  cfg.GroupID,
			Dialer:   dialer, // Authentification SASL/TLS
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

// worker : exécute la consommation Kafka pour un worker
func (c *Consumer[T]) worker(ctx context.Context, handle func(context.Context, T) error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done(): // Vérifie si le contexte est annulé
			log.Println("Arrêt du worker Kafka")
			return

		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Erreur de lecture Kafka : %v", err)
				continue // On continue pour garder la connexion active
			}

			// Instancier dynamiquement `T` sans factory
			var event T
			schema := event.GetSchema()

			// Décodage Avro
			decoder, err := avro.NewDecoder(schema, bytes.NewReader(msg.Value))
			if err != nil {
				log.Printf("Erreur chargement décodeur Avro : %v", err)
				continue
			}

			if err = decoder.Decode(&event); err != nil {
				log.Printf("Erreur de décodage Avro : %v", err)
				continue
			}

			// Exécuter la fonction métier
			if err = handle(ctx, event); err != nil {
				log.Printf("Erreur dans handle : %v", err)
			}
		}
	}
}

// Consume démarre plusieurs workers et écoute Kafka
func (c *Consumer[T]) Consume(ctx context.Context, cfg Config, handle func(context.Context, T) error) {
	var wg sync.WaitGroup

	for i := 0; i < cfg.NumWorkers; i++ {
		wg.Add(1)
		go c.worker(ctx, handle, &wg)
	}

	wg.Wait() // Attend que tous les workers terminent
}

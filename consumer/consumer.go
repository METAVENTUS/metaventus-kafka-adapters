package consumer

import (
	"bytes"
	"context"
	"crypto/tls"
	"log"
	"sync"
	"time"

	"github.com/METAVENTUS/metaventus-kafka-adapters/models"
	"github.com/hamba/avro"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Consumer générique Kafka
type Consumer[T models.AvroEvent] struct {
	cfg             Config
	reader          *kafka.Reader
	isBusinessHours bool

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewConsumer initialise un Kafka Consumer générique avec un type `T`
func NewConsumer[T models.AvroEvent](cfg Config) *Consumer[T] {
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

	conn, err := kafka.Dial("tcp", cfg.Brokers[0])
	if err != nil {
		log.Fatalf("❌ Échec de connexion à Kafka: %v", err)
	}
	defer conn.Close()
	log.Println("✅ Connexion réussie à Kafka")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		Dialer:   dialer, // Authentification SASL/TLS
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer[T]{
		cfg:             cfg,
		reader:          reader,
		isBusinessHours: cfg.IsBusinessHours,
	}
}

// Start lance la consommation dans une goroutine
// et démarre les workers en parallèle.
func (c *Consumer[T]) Start(ctx context.Context, handle func(context.Context, T) error) {
	workerContext, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	// On démarre N workers
	for i := 0; i < c.cfg.NumWorkers; i++ {
		c.wg.Add(1)
		go c.worker(workerContext, handle)
	}
}

// Close arrête la consommation
// - Annule le contexte pour que les workers s'arrêtent
// - Attend que tous les workers finissent
// - Ferme le reader Kafka
func (c *Consumer[T]) Close() error {
	// Annule le contexte pour que les goroutines worker s'arrêtent
	c.cancel()

	// Attend la fin de tous les workers
	c.wg.Wait()

	// Ferme le reader
	return c.reader.Close()
}

// worker : exécute la consommation Kafka pour un worker
func (c *Consumer[T]) worker(ctx context.Context, handle func(context.Context, T) error) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done(): // Vérifie si le contexte est annulé
			log.Println("Arrêt du worker Kafka (ctx.Done)")
			return

		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Erreur de lecture Kafka : %v", err)
				continue
			}

			// Vérifier si on est dans la plage horaire ou non (skip si hors plage)
			if c.isBusinessHours && !isBusinessHours(time.Now()) {
				log.Println("Hors plage horaire : skip le message temporairement")
				time.Sleep(1 * time.Minute)
				continue
			}

			var event T
			decoder, err := avro.NewDecoder(event.GetSchema(), bytes.NewReader(msg.Value))
			if err != nil {
				log.Printf("Erreur chargement décodeur Avro : %v", err)
				continue
			}

			if err = decoder.Decode(&event); err != nil {
				log.Printf("Erreur de décodage Avro : %v", err)
				continue
			}

			if err = handle(ctx, event); err != nil {
				log.Printf("Erreur dans handle : %v", err)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Fonctions utilitaires
// -----------------------------------------------------------------------------

// isBusinessHours indique si l'heure actuelle est dans la plage autorisée
// Ex. lundi-vendredi, 9h-19h (19h exclu)
func isBusinessHours(t time.Time) bool {
	if !isBusinessDay(t) {
		return false
	}
	hour := t.Hour()
	return hour >= 9 && hour < 19
}

// isBusinessDay vérifie si on est un jour de semaine (samedi/dimanche exclus)
func isBusinessDay(t time.Time) bool {
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	return true
}

package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/METAVENTUS/metaventus-kafka-adapters/consumer"
	"github.com/METAVENTUS/metaventus-kafka-adapters/producer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// TestEvent : modèle de test Avro
// Simule un message Kafka en Avro

type TestEvent struct {
	ID   string `avro:"id"`
	Data string `avro:"data"`
}

func (TestEvent) GetSchema() string {
	return `{
		"type": "record",
		"name": "TestEvent",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "data", "type": "string"}
		]
	}`
}

func (t TestEvent) PartitionKey() string { return t.ID }

// ConsumerIntegrationSuite : Suite de test

type ConsumerIntegrationSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	consumerContainer tc.Container
	kafkaHostPort     string // Adresse Kafka (host:port)
	topicName         string
}

// SetupSuite : Démarre Kafka et crée un topic
func (s *ConsumerIntegrationSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Démarrage du conteneur Kafka
	kafkaContainer, err := kafkacontainer.Run(s.ctx, "confluentinc/confluent-local:7.5.0",
		kafkacontainer.WithClusterID("test-cluster"),
	)
	s.Require().NoError(err, "Échec du démarrage du conteneur Kafka")

	// Récupération des infos du conteneur
	mappedPort, err := kafkaContainer.MappedPort(s.ctx, "9093")
	s.Require().NoError(err)
	host, err := kafkaContainer.Host(s.ctx)
	s.Require().NoError(err)
	s.kafkaHostPort = fmt.Sprintf("%s:%d", host, mappedPort.Int())

	log.Printf("Kafka démarré sur %s\n", s.kafkaHostPort)

	// Attendre que Kafka soit prêt
	s.Require().NoError(waitForKafkaReady(s.kafkaHostPort, 10), "Kafka n'est pas prêt")

	// Création d'un topic de test
	s.topicName = "test-topic"
	s.Require().NoError(createTopic(s.kafkaHostPort, s.topicName), "Échec de la création du topic")
}

// TearDownSuite : Arrête Kafka
func (s *ConsumerIntegrationSuite) TearDownSuite() {
	s.cancel()
	if s.consumerContainer != nil {
		_ = s.consumerContainer.Terminate(s.ctx)
	}
}

// Test_ConsumerReceivesMessage : Vérifie que le consumer reçoit bien un message
func (s *ConsumerIntegrationSuite) Test_ConsumerReceivesMessage() {
	cfg := consumer.Config{
		Brokers:         []string{s.kafkaHostPort},
		Topic:           s.topicName,
		GroupID:         "test-topic-cool",
		NumWorkers:      1,
		IsBusinessHours: false,
	}

	c := consumer.NewConsumer[TestEvent](cfg)
	msgCh := make(chan TestEvent, 1)

	// Démarrer la consommation dans une goroutine
	c.Start(s.ctx, func(ctx context.Context, event TestEvent) error {
		msgCh <- event
		return nil
	})
	defer func() { s.Require().NoError(c.Close()) }()

	// Attendre que le consumer soit bien en écoute
	time.Sleep(500 * time.Millisecond)

	// Publier un message de test
	testMsg := TestEvent{ID: "123", Data: "Hello Kafka"}
	producerCfg := producer.Config{
		Brokers: []string{s.kafkaHostPort},
		Topic:   s.topicName,
	}

	p, err := producer.NewProducer(s.ctx, producerCfg)
	s.Require().NoError(err, "Échec de l'initialisation du producer")
	s.Require().NoError(p.Publish(s.ctx, testMsg), "Échec de l'envoi du message")

	// Vérifier la réception du message avec timeout
	select {
	case evt := <-msgCh:
		s.T().Log(evt)
		s.Assert().Equal("123", evt.ID)
		s.Assert().Equal("Hello Kafka", evt.Data)
		return
	case <-time.After(200 * time.Second):
		s.Fail("Aucun message reçu après 10s")
	}
}

// createTopic : Crée un topic Kafka via segmentio/kafka-go
func createTopic(kafkaURL, topic string) error {
	conn, err := kafka.Dial("tcp", kafkaURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

// waitForKafkaReady : Vérifie que Kafka est bien accessible avant de lancer les tests
func waitForKafkaReady(kafkaURL string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		conn, err := kafka.Dial("tcp", kafkaURL)
		if err == nil {
			conn.Close()
			return nil
		}
		log.Printf("Kafka n'est pas prêt... retry dans 1s (%d/%d)", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Kafka n'a pas démarré après %d essais", maxRetries)
}

// Exécuter la suite de tests
func TestConsumerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ConsumerIntegrationSuite))
}

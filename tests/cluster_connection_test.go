package tests

import (
	"context"
	"fmt"
	"github.com/METAVENTUS/metaventus-kafka-adapters/consumer"
	"github.com/METAVENTUS/metaventus-kafka-adapters/models"
	"github.com/METAVENTUS/metaventus-kafka-adapters/producer"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
	"time"
)

func Test_produce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	consumerCfg, producerCfg := loadConfs()
	// load consumer
	testConsumer := consumer.NewConsumer[models.ModelExample](consumerCfg)
	testConsumer.Start(ctx, handle)

	testProducer, err := producer.NewProducer(ctx, producerCfg)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		event := models.ModelExample{
			ID:    uuid.NewString(),
			Email: fmt.Sprintf("my-email-is-%d@test.com", i),
			Name:  fmt.Sprintf("my-name-is-%d", i),
		}

		assert.NoError(t, testProducer.Publish(ctx, event))
	}

	<-time.After(time.Second * 20)
	cancel()
}

func handle(ctx context.Context, event models.ModelExample) error {
	fmt.Println(event)
	return nil
}

func loadConfs() (consumer.Config, producer.Config) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return consumerConf(), producerConf()
}

func consumerConf() consumer.Config {
	return consumer.Config{
		GroupID:         "test-test-id",
		Topic:           "test-topic",
		NumWorkers:      1,
		Brokers:         []string{os.Getenv("CONFLUENT_BOOTSTRAP_SERVERS")},
		Username:        os.Getenv("CONFLUENT_API_KEY"),
		Password:        os.Getenv("CONFLUENT_API_SECRET"),
		SASL:            true,
		TLS:             true,
		IsBusinessHours: false,
	}
}

func producerConf() producer.Config {
	return producer.Config{
		Topic:    "test-topic",
		Brokers:  []string{os.Getenv("CONFLUENT_BOOTSTRAP_SERVERS")},
		Username: os.Getenv("CONFLUENT_API_KEY"),
		Password: os.Getenv("CONFLUENT_API_SECRET"),
		SASL:     true,
		TLS:      true,
	}
}

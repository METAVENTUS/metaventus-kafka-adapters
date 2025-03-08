package main

import (
	"fmt"
	"log"
	"os"

	"github.com/METAVENTUS/metaventus-kafka-adapters/avro_kafka_config"
)

func main() {
	cfg := avro_kafka_config.LoadConfig()
	client := avro_kafka_config.NewKafkaClient(cfg)

	switch os.Args[1] {
	case "list-topics":
		topics, err := client.ListTopics()
		if err != nil {
			log.Fatalf("Erreur : %v", err)
		}
		fmt.Println("Topics Kafka :", topics)

	case "create-topic":
		if len(os.Args) < 3 {
			log.Fatalf("Usage : create-topic <topic-name>")
		}
		topic := os.Args[2]
		err := client.CreateTopic(topic, 3)
		if err != nil {
			log.Fatalf("Erreur : %v", err)
		}

	case "register-schema":
		if len(os.Args) < 3 {
			log.Fatalf("Usage : register-schema <schema-name>")
		}
		schema := os.Args[2]
		err := avro_kafka_config.RegisterSchema(cfg, schema)
		if err != nil {
			log.Fatalf("Erreur : %v", err)
		}

	default:
		fmt.Println("Commandes disponibles : list-topics, create-topic, register-schema")
	}
}

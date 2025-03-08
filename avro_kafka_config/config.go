package avro_kafka_config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config contient la configuration de connexion à Confluent Cloud
type Config struct {
	BootstrapServers     string
	APIKey               string
	APISecret            string
	SchemaRegistryURL    string
	SchemaRegistryKey    string
	SchemaRegistrySecret string
}

// LoadConfig charge la configuration depuis le fichier .env dans `cmd/`
func LoadConfig() Config {
	// Définir le chemin vers `.env` dans `cmd/`
	envPath := filepath.Join("cmd", ".env")

	// Charger le fichier `.env`
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Erreur lors du chargement du fichier .env (%s) : %v", envPath, err)
	}

	return Config{
		BootstrapServers:     os.Getenv("CONFLUENT_BOOTSTRAP_SERVERS"),
		APIKey:               os.Getenv("CONFLUENT_API_KEY"),
		APISecret:            os.Getenv("CONFLUENT_API_SECRET"),
		SchemaRegistryURL:    os.Getenv("CONFLUENT_SCHEMA_REGISTRY_URL"),
		SchemaRegistryKey:    os.Getenv("CONFLUENT_SCHEMA_REGISTRY_KEY"),
		SchemaRegistrySecret: os.Getenv("CONFLUENT_SCHEMA_REGISTRY_SECRET"),
	}
}

// Print affiche la configuration actuelle
func (c Config) Print() {
	fmt.Printf("Bootstrap Servers: %s\n", c.BootstrapServers)
	fmt.Printf("API Key: %s\n", c.APIKey)
	fmt.Printf("Schema Registry: %s\n", c.SchemaRegistryURL)
}

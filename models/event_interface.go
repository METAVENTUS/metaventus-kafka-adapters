package models

// AvroEvent définit l'interface pour tous les événements Kafka
type AvroEvent interface {
	GetSchema() string    // Retourne le schéma Avro
	PartitionKey() string // Retourne la clé de partition (optionnelle)
}

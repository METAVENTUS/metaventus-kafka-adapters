package models

import "github.com/METAVENTUS/metaventus-kafka-adapters/avro_schemas/schemas"

// ModelExample représente un événement Kafka
type ModelExample struct {
	ID    string `avro:"id"`
	Email string `avro:"email"`
	Name  string `avro:"name"`
}

// GetSchema retourne le schéma Avro associé
func (ModelExample) GetSchema() string {
	return schemas.ExampleSchema
}

// PartitionKey retourne la clé de partition (exemple : ID de l'utilisateur)
func (u ModelExample) PartitionKey() string {
	return u.ID
}

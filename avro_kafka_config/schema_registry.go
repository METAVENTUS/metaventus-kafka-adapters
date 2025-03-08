package avro_kafka_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	avroschemas "github.com/METAVENTUS/metaventus-kafka-adapters/avro_schemas"
	"io/ioutil"
	"net/http"
)

// SchemaRequest représente le format d'une requête pour ajouter un schéma Avro
type SchemaRequest struct {
	Schema string `json:"schema"`
}

// RegisterSchema enregistre un schéma Avro dans Confluent Schema Registry
func RegisterSchema(cfg Config, schemaName string) error {
	schema, exists := avroschemas.AvroSchemas[schemaName]
	if !exists {
		return fmt.Errorf("schéma introuvable pour %s", schemaName)
	}

	url := fmt.Sprintf("%s/subjects/%s-value/versions", cfg.SchemaRegistryURL, schemaName)

	data, err := json.Marshal(SchemaRequest{Schema: schema})
	if err != nil {
		return fmt.Errorf("erreur de conversion JSON : %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête : %w", err)
	}

	req.SetBasicAuth(cfg.SchemaRegistryKey, cfg.SchemaRegistrySecret)
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi de la requête : %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("échec de l'enregistrement du schéma : %s", string(body))
	}

	fmt.Printf("Schéma %s enregistré avec succès\n", schemaName)
	return nil
}

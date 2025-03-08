package consumer

// Config pour le Consumer Kafka
type Config struct {
	Brokers    []string `json:"brokers"`     // Liste des brokers Kafka
	Topic      string   `json:"topic"`       // Nom du topic Kafka
	GroupID    string   `json:"group_id"`    // ID du groupe de consommation
	NumWorkers int      `json:"num_workers"` // Nombre de workers concurrents
	Username   string   `json:"username"`    // Identifiant Kafka
	Password   string   `json:"password"`    // Mot de passe Kafka
	SASL       bool     `json:"sasl"`        // Activer l'authentification SASL
	TLS        bool     `json:"tls"`         // Activer TLS (si nécessaire)
}

package producer

// Config pour le Producer Kafka
type Config struct {
	Brokers  []string `json:"brokers"`  // Liste des brokers Kafka
	Topic    string   `json:"topic"`    // Nom du topic Kafka
	Username string   `json:"username"` // Identifiant Kafka (si authentification)
	Password string   `json:"password"` // Mot de passe Kafka (si authentification)
	SASL     bool     `json:"sasl"`     // Activer l'authentification SASL
	TLS      bool     `json:"tls"`      // Activer TLS (si nécessaire)
}

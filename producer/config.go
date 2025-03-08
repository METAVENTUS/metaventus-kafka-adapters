package producer

import (
	"os"
	"strconv"
	"strings"
)

// Config pour le Producer Kafka
type Config struct {
	Brokers  []string // Liste des brokers Kafka
	Topic    string   // Nom du topic Kafka
	Username string   // Identifiant Kafka (si authentification)
	Password string   // Mot de passe Kafka (si authentification)
	SASL     bool     // Activer l'authentification SASL
	TLS      bool     // Activer TLS (si nécessaire)
}

// LoadConfigFromEnv construit la Config en lisant les variables d'environnement
func LoadConfigFromEnv() Config {
	// Ex: PRODUCER_BROKERS="broker1:9092,broker2:9092"
	var brokers []string
	if brokersEnv := os.Getenv("PRODUCER_BROKERS"); brokersEnv != "" {
		brokers = strings.Split(brokersEnv, ",")
	}

	sasl := parseBool(os.Getenv("PRODUCER_SASL"))
	tls := parseBool(os.Getenv("PRODUCER_TLS"))

	cfg := Config{
		Brokers:  brokers,
		Topic:    os.Getenv("PRODUCER_TOPIC"),
		Username: os.Getenv("PRODUCER_USERNAME"),
		Password: os.Getenv("PRODUCER_PASSWORD"),
		SASL:     sasl,
		TLS:      tls,
	}
	return cfg
}

// parseBool convertit une chaîne en bool (renvoie false si vide ou inconnu)
func parseBool(val string) bool {
	switch strings.ToLower(val) {
	case "true", "1", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// parseInt convertit une chaîne en int, renvoyant defaultVal en cas d'erreur ou vide
func parseInt(val string, defaultVal int) int {
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

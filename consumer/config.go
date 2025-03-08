package consumer

import (
	"os"
	"strconv"
	"strings"
)

// Config pour le Consumer Kafka
type Config struct {
	Brokers         []string
	Topic           string
	GroupID         string
	NumWorkers      int
	Username        string
	Password        string
	SASL            bool
	TLS             bool
	IsBusinessHours bool
}

// LoadConfigFromEnv construit la Config en lisant les variables d'environnement
func LoadConfigFromEnv() Config {
	// Par exemple, on définit les brokers comme une liste séparée par des virgules
	var brokers []string
	if brokersEnv := os.Getenv("KAFKA_BROKERS"); brokersEnv != "" {
		brokers = strings.Split(brokersEnv, ",")
	}

	// Conversion string → bool
	sasl := parseBool(os.Getenv("KAFKA_SASL"))
	tls := parseBool(os.Getenv("KAFKA_TLS"))
	isBusinessHours := parseBool(os.Getenv("KAFKA_IS_BUSINESS_HOURS"))

	// Conversion string → int
	numWorkers := parseInt(os.Getenv("KAFKA_NUM_WORKERS"), 1)

	cfg := Config{
		Brokers:         brokers,
		Topic:           os.Getenv("KAFKA_TOPIC"),
		GroupID:         os.Getenv("KAFKA_GROUP_ID"),
		NumWorkers:      numWorkers,
		Username:        os.Getenv("KAFKA_USERNAME"),
		Password:        os.Getenv("KAFKA_PASSWORD"),
		SASL:            sasl,
		TLS:             tls,
		IsBusinessHours: isBusinessHours,
	}
	return cfg
}

// parseBool convertit une chaîne en bool (false si vide ou erreur)
func parseBool(val string) bool {
	switch strings.ToLower(val) {
	case "true", "1", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// parseInt convertit une chaîne en int, en renvoyant defaultVal en cas d'erreur
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

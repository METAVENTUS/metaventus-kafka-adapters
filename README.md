# Metaventus Kafka Adapters

Bienvenue dans **Metaventus Kafka Adapters** – un **monorepo** Go qui fournit tous les éléments nécessaires pour interagir avec un **cluster Kafka** hébergé sur **Confluent Cloud**. L’objectif est de proposer des **clients Producer et Consumer** robustes, des **modèles Avro** clairement définis et un **mini-CLI** en Go pour gérer la configuration Confluent (topics, schémas, etc.).

---

## 📌 Sommaire

1. [Structure générale du projet](#1-structure-générale-du-projet)
2. [Fonctionnalités](#2-fonctionnalités)
3. [Configuration & Fichier `.env`](#3-configuration--fichier-env)
4. [Modules principaux](#4-modules-principaux)
5. [Exemple de consommation/production Kafka](#5-exemple-de-consommationproduction-kafka)
6. [Exemple de gestion Confluent Cloud via CLI](#6-exemple-de-gestion-confluent-cloud-via-cli)
7. [Tests et intégration continue](#7-tests-et-intégration-continue)
8. [Dépendances et outils utilisés](#8-dépendances-et-outils-utilisés)

---

## 1. Structure générale du projet

```
metaventus-kafka-adapters/
│── avro_kafka_config/         # Gestion de Confluent Cloud (topics, schémas) + CLI
│   ├── config.go              # Chargement de la config (via .env)
│   ├── client.go              # Méthodes pour créer/lister les topics
│   ├── schema_registry.go     # Enregistrement des schémas Avro auprès du Schema Registry
│   ├── README.md              # Documentation spécifique au module avro_kafka_config
│
│── avro_schemas/              # Centralisation des schémas Avro
│   ├── avro_schemas.go        # Map { nomDuSchéma : JSON du schéma }
│   ├── schemas/               # Dossier contenant les fichiers de schémas Avro
│
│── cmd/                       # CLI pour exécuter des commandes (ex: créer un topic, etc.)
│   ├── main.go                # Programme principal pour gérer Confluent Cloud
│   ├── .env                   # Fichier de configuration locale (Non commité)
│
│── consumer/                  # Package pour implémenter un Consumer Kafka générique
│   ├── config.go              # Config Consumer (brokers, group ID, nombre de workers...)
│   ├── consumer.go            # Implémentation d'un Consumer générique (génériques Go)
│
│── producer/                  # Package pour implémenter un Producer Kafka générique
│   ├── config.go              # Config Producer (brokers, authentification SASL/TLS...)
│   ├── producer.go            # Implémentation d'un Producer générique
│
│── models/                    # Modèles Go correspondant aux schémas Avro
│
│── go.mod                     # Module Go principal
│── README.md                  # Documentation générale du repository
```

---

## 2. Fonctionnalités

- **Producer/Consumer Kafka** :
   - Structuration claire (config, encodage Avro, clés de partition, etc.).
   - Possibilité de consommer plusieurs messages en parallèle (multi-workers).
   - Gestion de l’authentification SASL/TLS pour Confluent Cloud.
   - **Filtrage optionnel des messages en fonction des horaires d'ouverture**.

- **Schemas Avro centralisés** :
   - Tous les schémas Avro sont stockés dans `avro_schemas/`.
   - Chaque modèle Go (`models/`) peut renvoyer son schéma via `GetSchema()`.
   - Compatible avec un usage direct du **Confluent Schema Registry**.

- **Gestion avancée du Consumer** :
   - Vérification de la connexion à Kafka au démarrage.
   - Gestion dynamique des **heures d’ouverture** (peut être activée via `IsBusinessHours`).
   - Support du **multi-workers** pour la consommation parallèle.

- **CLI pour Confluent Cloud** :
   - Création et listing de topics.
   - Enregistrement des schémas Avro.
   - Chargement automatique des credentials via `.env`.

---

## 3. Configuration & Fichier `.env`

Ajout d'un paramètre optionnel pour gérer les **heures d'ouverture du Consumer** :

```ini
# Activer la gestion des heures d'ouverture (true/false)
KAFKA_CONSUMER_BUSINESS_HOURS=true
```

Si activé, le consumer ne traitera que les messages entre **9h et 19h, du lundi au vendredi**.

---

## 4. Exemple de consommation avec gestion des horaires

```go
cfg := consumer.Config{
Brokers:        []string{"your-cluster.confluent.cloud:9092"},
Topic:          "example-topic",
GroupID:        "example-group",
NumWorkers:     3,
IsBusinessHours: true, // ← Activer la gestion des horaires
}

c := consumer.NewConsumer[models.ModelExample](cfg)

c.Consume(context.Background(), cfg, func(ctx context.Context, event models.ModelExample) error {
log.Printf("Message reçu : %+v", event)
return nil
})
```

---

## 5. Exemple de gestion Confluent Cloud via CLI

1. **Lister les topics** :
   ```sh
go run cmd/main.go list-topics
```
2. **Créer un topic** :
   ```sh
go run cmd/main.go create-topic my-topic
```
3. **Enregistrer un schéma** :
   ```sh
go run cmd/main.go register-schema ExampleSchema
```

---

## 6. Tests et intégration continue

Ce projet inclut des tests d'intégration pour garantir le bon fonctionnement des **Consumers** et **Producers**.

### 🔍 **Tests d'intégration avec TestContainers**
Nous utilisons **TestContainers** pour lancer un Kafka temporaire et tester nos composants en conditions réelles.

Exemple de test de consommation Kafka :

```go
func (s *ConsumerIntegrationSuite) Test_ConsumerReceivesMessage() {
cfg := consumer.Config{
Brokers:        []string{s.kafkaHostPort},
Topic:          s.topicName,
GroupID:        "test-group",
NumWorkers:     1,
IsBusinessHours: true,
}

c := consumer.NewConsumer[TestEvent](cfg)
msgCh := make(chan TestEvent, 1)

c.Start(s.ctx, func(ctx context.Context, event TestEvent) error {
msgCh <- event
return nil
})
defer func() { s.Require().NoError(c.Close()) }()

testMsg := TestEvent{ID: "123", Data: "Hello Kafka"}
p, err := producer.NewProducer(s.ctx, producer.Config{
Brokers: []string{s.kafkaHostPort},
Topic:   s.topicName,
})
s.Require().NoError(err)
s.Require().NoError(p.Publish(s.ctx, testMsg))

select {
case evt := <-msgCh:
s.Assert().Equal("123", evt.ID)
s.Assert().Equal("Hello Kafka", evt.Data)
case <-time.After(10 * time.Second):
s.Fail("Aucun message reçu après 10s")
}
}
```

### 🔧 **Exécution des tests**
Pour exécuter les tests :
```sh
go test ./...
```

Ou pour afficher plus de détails :
```sh
go test -v ./...
```

---

## 7. Dépendances et outils utilisés

- **[segmentio/kafka-go](https://github.com/segmentio/kafka-go)**
- **[hamba/avro](https://github.com/hamba/avro)**
- **[testcontainers-go](https://github.com/testcontainers/testcontainers-go)**
- **API REST** de **Confluent Schema Registry**
- **Go 1.21+** pour les génériques et améliorations de performance.

---

C'est à jour avec toutes les nouvelles fonctionnalités ! 🚀 Dis-moi si tu veux d'autres ajustements.
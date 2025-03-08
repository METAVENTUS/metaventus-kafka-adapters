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
7. [Dépendances et outils utilisés](#7-dépendances-et-outils-utilisés)  
8. [Contribuer](#8-contribuer)

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
│   │   ├── example.go         # Exemples de schémas Avro
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
│   ├── model_example.go       # Exemple de modèle (associe un schéma Avro)
│
│── go.mod                     # Module Go principal
│── README.md                  # Documentation générale du repository
```

---

## 2. Fonctionnalités

1. **Producer/Consumer Kafka** :  
   - Structuration claire (config, encodage Avro, clés de partition, etc.).
   - Possibilité de consommer plusieurs messages en parallèle (multi-workers).
   - Gestion de l’authentification SASL/TLS pour Confluent Cloud.

2. **Schemas Avro centralisés** :  
   - Tous les schémas Avro sont stockés dans `avro_schemas/`.
   - Chaque modèle Go (`models/`) peut renvoyer son schéma via `GetSchema()`.
   - Compatible avec un usage direct du **Confluent Schema Registry**.

3. **CLI pour Confluent Cloud** :  
   - Création et listing de topics.  
   - Enregistrement des schémas Avro.  
   - Chargement automatique des credentials via `.env`.

4. **Monorepo** :
   - Un seul repository gère la **configuration**, les **clients Kafka**, les **schémas Avro** et le **mini-CLI**.  
   - Simplifie la maintenance et le déploiement.

---

## 3. Configuration & Fichier `.env`

Le fichier `.env` (non commité) se situe dans le dossier `cmd/`. Il doit contenir :

```ini
# Configuration Kafka
CONFLUENT_BOOTSTRAP_SERVERS=your-cluster.confluent.cloud:9092
CONFLUENT_API_KEY=your-api-key
CONFLUENT_API_SECRET=your-api-secret

# Configuration Schema Registry
CONFLUENT_SCHEMA_REGISTRY_URL=https://your-schema-registry-url
CONFLUENT_SCHEMA_REGISTRY_KEY=your-schema-registry-key
CONFLUENT_SCHEMA_REGISTRY_SECRET=your-schema-registry-secret
```

**Important** : Ajoutez `cmd/.env` dans votre `.gitignore` pour éviter de commiter vos identifiants sensibles.

---

## 4. Modules principaux

1. **`avro_kafka_config`**
   - **`client.go`** : créer/lister des topics sur Kafka.
   - **`schema_registry.go`** : enregistrer les schémas Avro auprès du Confluent Schema Registry.

2. **`avro_schemas`**
   - **`avro_schemas.go`** : Map `{ schemaName : schemaDefinition }`.
   - **`schemas/`** : Fichiers contenant le JSON Avro pour chaque schéma (ex: `example.go`).

3. **`consumer/`**
   - Permet de démarrer un Consumer Kafka générique (avec ou sans multi-workers).
   - Gère l’authentification SASL/TLS pour se connecter à Confluent Cloud.

4. **`producer/`**
   - Permet de démarrer un Producer Kafka générique.
   - Encode automatiquement les messages en Avro avant l’envoi.
   - Gère l’authentification SASL/TLS.

5. **`models/`**
   - Les structs Go représentant chaque message Kafka, qui renvoient leur **schéma Avro** via `GetSchema()`.

6. **`cmd/`**
   - **`main.go`** : CLI principal qui utilise `avro_kafka_config` pour lister, créer des topics ou enregistrer un schéma.
   - **`.env`** : Variables d’environnement pour Confluent Cloud.

---

## 5. Exemple de consommation/production Kafka

### Exemple d’un Producer minimal

```go
package main

import (
  "context"
  "log"

  "github.com/METAVENTUS/metaventus-kafka-adapters/producer"
  "github.com/METAVENTUS/metaventus-kafka-adapters/models"
)

func main() {
  // Configurer le Producer
  cfg := producer.Config{
    Brokers:  []string{"your-cluster.confluent.cloud:9092"},
    Topic:    "example-topic",
    Username: "your-api-key",
    Password: "your-api-secret",
    SASL:     true,
    TLS:      true,
  }

  // Initialiser le Producer
  p := producer.NewProducer(cfg)

  // Publier un événement (Avro)
  event := models.ModelExample{ID: "123", Email: "test@example.com", Name: "John"}
  if err := p.Publish(context.Background(), event); err != nil {
    log.Fatalf("Erreur publication : %v", err)
  }
}
```

### Exemple d’un Consumer minimal

```go
package main

import (
  "context"
  "log"

  "github.com/METAVENTUS/metaventus-kafka-adapters/consumer"
  "github.com/METAVENTUS/metaventus-kafka-adapters/models"
)

func main() {
  // Configurer le Consumer
  cfg := consumer.Config{
    Brokers:    []string{"your-cluster.confluent.cloud:9092"},
    Topic:      "example-topic",
    GroupID:    "example-group",
    Username:   "your-api-key",
    Password:   "your-api-secret",
    SASL:       true,
    TLS:        true,
    NumWorkers: 3, // consommer en parallèle
  }

  // Initialiser le Consumer
  c := consumer.NewConsumer[models.ModelExample](cfg)

  // Consommer les messages
  c.Consume(context.Background(), cfg, func(ctx context.Context, event models.ModelExample) error {
    log.Printf("Message reçu : %+v", event)
    return nil
  })
}
```

---

## 6. Exemple de gestion Confluent Cloud via CLI

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

Les schémas sont définis dans `avro_schemas/schemas/` et référencés par la map globale `AvroSchemas` dans `avro_schemas/avro_schemas.go`.

---

## 7. Dépendances et outils utilisés

- **[segmentio/kafka-go](https://github.com/segmentio/kafka-go)** : Gestion Kafka (topic, production, consommation).
- **[joho/godotenv](https://github.com/joho/godotenv)** : Chargement du fichier `.env`.
- **API REST** de **Confluent Schema Registry** : Pour enregistrer et gérer les schémas Avro.
- **Go 1.21+** : Pour profiter des **génériques** (Consumer générique, etc.).

---


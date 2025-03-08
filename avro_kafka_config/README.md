# Présentation de `avro_kafka_config`

`avro_kafka_config` est un module Go destiné à faciliter la **gestion d’un cluster Kafka Confluent Cloud** à travers :
1. La **création et l’inventaire des topics** (via l’API Kafka).
2. L’**enregistrement des schémas Avro** dans le **Confluent Schema Registry** (via l’API Schema Registry).

L’objectif est de fournir un **mini-CLI** en Go, capable de gérer des configurations complexes depuis un fichier `.env`, sans avoir à manipuler manuellement des scripts externes ou des commandes CLI spécifiques à Confluent.

---

## 1. Architecture générale

Le code est organisé de manière à :
- **Charger la configuration Confluent** depuis un fichier `.env` (stocké dans `cmd/`).  
- **Se connecter à Confluent Cloud** grâce à des identifiants `API Key` / `API Secret`.  
- **Lister et créer des topics** Kafka via des méthodes du package `client.go`.  
- **Enregistrer des schémas Avro** dans Confluent Schema Registry via des méthodes du package `schema_registry.go`.  
- **Centraliser les définitions de schémas Avro** au sein du répertoire `avro_schemas/`, afin que les modèles et le CLI utilisent exactement les mêmes schémas (aucune duplication).

**Arborescence indicative** :

```
metaventus-kafka-adapters/
│── avro_kafka_config/        
│   ├── config.go             # Lecture de la config depuis le .env
│   ├── client.go             # Gestion des topics Kafka
│   ├── schema_registry.go    # Gestion des schémas Avro (Schema Registry)
│── avro_schemas/             
│   ├── avro_schemas.go       # Map { nomDuSchéma : JSON du schéma }
│   ├── schemas/              # Fichiers définissant chaque schéma Avro
│   │   └── example.go        
│── cmd/                      
│   ├── main.go               # CLI principal
│   ├── .env                  # Fichier de configuration locale
│── consumer/                 
│── producer/                 
│── models/                   
│── go.mod
│── README.md
```

---

## 2. Composants principaux

### 2.1. Fichier `.env`

Situé dans le dossier `cmd/`, il contient toutes les informations de connexion à Confluent Cloud :

```ini
CONFLUENT_BOOTSTRAP_SERVERS=your-cluster.confluent.cloud:9092
CONFLUENT_API_KEY=your-api-key
CONFLUENT_API_SECRET=your-api-secret

CONFLUENT_SCHEMA_REGISTRY_URL=https://your-schema-registry-url
CONFLUENT_SCHEMA_REGISTRY_KEY=your-schema-registry-key
CONFLUENT_SCHEMA_REGISTRY_SECRET=your-schema-registry-secret
```

- **`CONFLUENT_BOOTSTRAP_SERVERS`** : l’adresse du broker Kafka Confluent.
- **`CONFLUENT_API_KEY` / `CONFLUENT_API_SECRET`** : identifiants pour la connexion SASL/PLAIN.
- **`CONFLUENT_SCHEMA_REGISTRY_URL`** : URL du Schema Registry.
- **`CONFLUENT_SCHEMA_REGISTRY_KEY` / `CONFLUENT_SCHEMA_REGISTRY_SECRET`** : identifiants de connexion au Schema Registry.

> **Note :** Ce fichier ne doit **jamais** être commité (ajouter `cmd/.env` à votre `.gitignore`).

---

### 2.2. Configuration (`config.go`)

Le fichier `config.go` permet de :
- **Charger** le `.env` grâce à la librairie [joho/godotenv](https://github.com/joho/godotenv).
- **Stocker** ces informations dans une struct `Config`.
- Les exploiter ensuite dans tout le projet (topics, schema registry, etc.).

---

### 2.3. Gestion des topics (`client.go`)

Ce fichier expose un **`KafkaClient`** qui :

- **Se connecte à Confluent Cloud** en utilisant les informations de `Config` et la lib [segmentio/kafka-go](https://github.com/segmentio/kafka-go).
- Offre des méthodes telles que :
    - `CreateTopic(topic string, partitions int)` : créer un topic avec le nombre de partitions indiqué.
    - `ListTopics()` : lister les topics existants dans le cluster.

Grâce à **kafka-go**, ces opérations s’effectuent **directement via l’API Kafka**.

---

### 2.4. Gestion des schémas Avro (`schema_registry.go`)

Pour enregistrer des **schémas Avro** dans le **Confluent Schema Registry**, le fichier `schema_registry.go` :

1. **Charge** le schéma Avro (en JSON) depuis la map `AvroSchemas` (définie dans `avro_schemas/avro_schemas.go`).
2. **Envoie** ce schéma au sujet approprié via un appel POST sur l’URL `/subjects/<schemaName>-value/versions`.
3. **Authentifie** la requête avec `CONFLUENT_SCHEMA_REGISTRY_KEY` et `CONFLUENT_SCHEMA_REGISTRY_SECRET`.

Les schémas sont donc **centralisés** dans un seul package (`avro_schemas`), ce qui évite les divergences entre :

- Les **modèles** Go (`models/`)
- Les **schémas** utilisés par Schema Registry
- Les **données** publiées ou consommées

---

## 3. Usage du CLI (`cmd/main.go`)

Un **mini-CLI** est disponible dans le fichier `cmd/main.go`. Il expose trois commandes principales :

1. **`list-topics`**  
   Liste tous les topics disponibles dans le cluster Kafka :
   ```sh
   go run cmd/main.go list-topics
   ```
2. **`create-topic <nom>`**  
   Crée un topic nommé `<nom>` avec (par défaut) 3 partitions, réplication factor = 3 :
   ```sh
   go run cmd/main.go create-topic mon-topic
   ```
3. **`register-schema <nom-du-schéma>`**  
   Enregistre le schéma Avro correspondant à `<nom-du-schéma>` (défini dans `avro_schemas/avro_schemas.go`) sur le Schema Registry :
   ```sh
   go run cmd/main.go register-schema ExampleSchema
   ```

Si la commande n’est pas reconnue, le CLI propose la liste des commandes disponibles.

---

## 4. Ajouter un nouveau schéma Avro

1. **Créer un fichier** dans `avro_schemas/schemas/` (par ex. `order_created.go`), où vous définissez les **constantes** représentant le nom et la définition JSON du schéma :
   ```go
   package schemas

   const OrderCreatedName = "OrderCreated"
   const OrderCreatedSchema = `{
     "type": "record",
     "name": "OrderCreated",
     "fields": [
       {"name": "order_id", "type": "string"},
       {"name": "user_id", "type": "string"},
       {"name": "total_price", "type": "float"}
     ]
   }`
   ```
2. **Référencer** ce nouveau schéma dans `avro_schemas/avro_schemas.go` :
   ```go
   package avroschemas

   import "github.com/METAVENTUS/metaventus-kafka-adapters/avro_schemas/schemas"

   var AvroSchemas = map[string]string{
     schemas.ExampleName:       schemas.ExampleSchema,
     schemas.OrderCreatedName:  schemas.OrderCreatedSchema,
   }
   ```
3. **Enregistrer** le schéma via la CLI :
   ```sh
   go run cmd/main.go register-schema OrderCreated
   ```

---

## 5. Exemple d’utilisation dans un modèle Go

Si vous avez un **modèle** qui représente votre événement Kafka, vous pouvez directement **retourner le schéma** depuis `GetSchema()` :

```go
package models

import "github.com/METAVENTUS/metaventus-kafka-adapters/avro_schemas/schemas"

type OrderCreated struct {
  OrderID    string  `avro:"order_id"`
  UserID     string  `avro:"user_id"`
  TotalPrice float64 `avro:"total_price"`
}

func (OrderCreated) GetSchema() string {
  return schemas.OrderCreatedSchema
}

func (o OrderCreated) PartitionKey() string {
  return o.OrderID
}
```

De cette manière, **producteurs** et **consommateurs** Go peuvent sérialiser/désérialiser l’événement en Avro en toute cohérence avec le **Schema Registry**.

---

## 6. Résumé des commandes

| Commande                        | Description                                             |
|--------------------------------|---------------------------------------------------------|
| **`list-topics`**              | Liste les topics Kafka existants                       |
| **`create-topic <nom>`**       | Crée un topic Kafka                                    |
| **`register-schema <nom>`**    | Enregistre un schéma Avro dans le Schema Registry      |

---

## 7. Dépendances

- [**segmentio/kafka-go**](https://github.com/segmentio/kafka-go) : interaction avec Kafka (création, lecture, listing de topics).
- [**joho/godotenv**](https://github.com/joho/godotenv) : chargement du fichier `.env`.
- **API REST** de Confluent Schema Registry : pour enregistrer les schémas Avro.

---

## 8. Dépannage

1. **`.env` introuvable**
    - Vérifier que le fichier `.env` est bien dans `cmd/.env`.
    - Si besoin, le déplacer :
      ```sh
      mv .env cmd/
      ```  

2. **Erreur d’authentification**
    - Vérifier les valeurs de `CONFLUENT_API_KEY` / `CONFLUENT_API_SECRET`.
    - Idem pour `CONFLUENT_SCHEMA_REGISTRY_KEY` / `CONFLUENT_SCHEMA_REGISTRY_SECRET`.

---

## 9. Conclusion

Grâce à `avro_kafka_config`, vous pouvez :

- **Créer** et **lister** vos topics Kafka Confluent Cloud.
- **Enregistrer** vos schémas Avro dans le Schema Registry.
- **Aligner** vos modèles Go avec ces mêmes schémas pour assurer la **compatibilité** de la sérialisation.

C’est une solution **simple, modulaire** et **scalable** pour gérer **Kafka** et **Confluent** dans un seul **monorepo Go**.

**Bon usage !** 

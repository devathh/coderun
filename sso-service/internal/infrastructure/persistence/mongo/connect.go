package mongodb

import (
	"context"
	"fmt"
	"net"

	"github.com/devathh/coderun/sso-service/internal/infrastructure/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func Connect(cfg *config.Config) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s", net.JoinHostPort(
		cfg.Secrets.Mongo.Host,
		cfg.Secrets.Mongo.Port,
	))

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with mongo: %w", err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	return client, nil
}

func Close(client *mongo.Client) error {
	return client.Disconnect(context.Background())
}

func CreateIndexes(client *mongo.Client) error {
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	idIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	coll := client.Database("coderun").Collection("users")

	if _, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{emailIndex, idIndex}); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

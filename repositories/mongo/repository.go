package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct {
	client    *mongo.Client
	itemsColl *mongo.Collection
}

func NewRepository(uri, dbName string) (*Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	itemsColl := client.Database(dbName).Collection(itemCollName)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "account_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := itemsColl.Indexes().CreateOne(ctx, indexModel); err != nil {
		return nil, fmt.Errorf("failed to create item name-account index: %w", err)
	}

	return &Repository{
		client:    client,
		itemsColl: itemsColl,
	}, nil
}

func (r *Repository) Close(ctx context.Context) error {
	if r.client == nil {
		return nil
	}
	return r.client.Disconnect(ctx)
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

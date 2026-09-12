package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient(
	ctx context.Context,
	mongoURL string,
) (*mongo.Client, error) {

	client, err := mongo.Connect(
		options.Client().ApplyURI(mongoURL),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create MongoDB client: %w",
			err,
		)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)

		return nil, fmt.Errorf(
			"failed to connect to MongoDB: %w",
			err,
		)
	}

	return client, nil
}

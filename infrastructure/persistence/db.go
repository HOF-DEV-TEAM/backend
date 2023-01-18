package persistence

import (
	"bitbucket.org/hofng/hofApp/domain/repository"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"time"
)

const (
	NAME = "HOF"
)

type mongoStore struct {
	mongoClient  *mongo.Client
	databaseName string
	logger       *zap.Logger
}

func New(connectUri, databaseName string, logger *zap.Logger) (repository.Repositories, *mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectUri))
	if err != nil {
		return nil, nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, err
	}
	return &mongoStore{mongoClient: client, databaseName: databaseName, logger: logger}, client, nil
}

func (repo *mongoStore) col(collectionName string) *mongo.Collection {
	return repo.mongoClient.Database(repo.databaseName).Collection(collectionName)
}

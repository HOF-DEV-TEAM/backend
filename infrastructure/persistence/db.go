package persistence

import (
	"bitbucket.org/hofng/hofApp/domain/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.uber.org/zap"
)

type mongoStore struct {
	mongoClient  *mongo.Client
	databaseName string
	logger       *zap.Logger
}

func New(client *mongo.Client, databaseName string, logger *zap.Logger) repository.Repositories {
	return &mongoStore{mongoClient: client, databaseName: databaseName, logger: logger}
}

func (repo *mongoStore) col(collectionName string) *mongo.Collection {
	return repo.mongoClient.Database(repo.databaseName, &options.DatabaseOptions{ReadConcern: readconcern.Local()}).Collection(collectionName)
}

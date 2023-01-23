package persistence

import (
	"bitbucket.org/hofng/hofApp/domain/entity"
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap/zaptest"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	hofDbName = "hof"
)

var (
	mongoDbPort = ""
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal(err)
	}

	resource, err := pool.Run("mongo", "4.2.9", []string{
		"MONGO_INITDB_DATABASE=" + hofDbName,
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	mongoDbPort = resource.GetPort("27017/tcp")
	if err := pool.Retry(func() error {
		_, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			panic(err)
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	code := m.Run()

	rand.Seed(time.Now().UnixNano())
	os.Exit(code)
}

func TestMongoStore_CreateUser(t *testing.T) {
	const (
		success = iota
	)
	var (
		user = entity.User{ID: "hjhdy839903bnnbx"}
	)
	var tests = []struct {
		name     string
		arg      entity.User
		testType int
	}{
		{
			name:     "Successfully created users",
			arg:      user,
			testType: success,
		},
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	repo := New(client, hofDbName, zaptest.NewLogger(t))
	_, err = client.Database(hofDbName).Collection(entity.UserCollectionName).InsertOne(context.Background(), user)
	if err != nil {
		assert.NoError(t, err)
		t.Fail()
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.testType {
			case success:
				repo.CreateUser(test.arg)
				assert.Equal(t, user, test.arg)
			}
		})
	}
}

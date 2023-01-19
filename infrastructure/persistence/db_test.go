package persistence

import (
	"bitbucket.org/hofng/hofApp/domain/entity"
	"bitbucket.org/hofng/hofApp/infrastructure/library/logger"
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
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
		var err error
		connectURL := fmt.Sprintf("mongodb://localhost:%s", mongoDbPort)
		_, _, err = New(connectURL, "hof", logger.New())
		if err != nil {
			return err
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

	connectURI := fmt.Sprintf("mongodb://localhost:%s", mongoDbPort)
	repo, client, errRt := New(connectURI, hofDbName, zaptest.NewLogger(t))
	if errRt != nil {
		fmt.Println(errRt)
	}
	assert.NotNil(t, client)

	_, err := client.Database(hofDbName).Collection(entity.UserCollectionName).InsertOne(context.Background(), user)
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

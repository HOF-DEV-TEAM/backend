package persistence

import (
	"bitbucket.org/hofng/hofApp/domain/entity"
	"bitbucket.org/hofng/hofApp/infrastructure/library/errorHandler"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

func (repo *mongoStore) CreateUser(user entity.User) {
	//TODO implement me
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println(user)
	_, err := repo.col("user").InsertOne(ctx, user)
	if err != nil {
		errorHandler.Format(errorHandler.DatabaseError, err)
	}
	repo.logger.Info("user created", zap.String("user_id", user.ID))
}

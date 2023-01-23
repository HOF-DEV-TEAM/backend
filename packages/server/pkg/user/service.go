package user

import (
	"context"

	"go.uber.org/zap"
)


type Service interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
}

type userService struct {
	repo Repository
	log *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &userService{log: log, repo: repo}
}

func(svc *userService) CreateUser(ctx context.Context, user *User) (*User, error) {
	return nil, nil
}
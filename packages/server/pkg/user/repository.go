package user

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)


type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
}

type userRepository struct {
	db 		*sql.DB
	log 	*zap.Logger
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &userRepository{db: db, log: logger}
}

func (r userRepository) Create(ctx context.Context, user *User) (*User, error) {
	return nil, nil
}
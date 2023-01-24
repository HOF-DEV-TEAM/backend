package user

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)


type Service interface {
	CreateUser(ctx context.Context, user *User) (*UserAndToken, error)
}

type userService struct {
	repo Repository
	log *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	return &userService{log: log, repo: repo}
}

func (s *userService) validateStruct(user *User) error {
	validate := validator.New()

	return validate.Struct(user)
}

func(svc *userService) CreateUser(ctx context.Context, user *User) (*UserAndToken, error) {

	err := svc.validateStruct(user)

	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "Email":
				return nil, errors.New("Email is required")
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = svc.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, errors.New("Invalid request")
	}

	// leading and trailing whitespaces
	user.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(user.Password))))


	result, err := svc.repo.Create(ctx, user)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg", 
			zap.String("method", "Create"), 
			zap.String("error", err.Error()), 
		)
		return nil, err
	}

	return &UserAndToken{User: result, Token:  ""}, nil 
}
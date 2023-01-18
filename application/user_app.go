package application

import "bitbucket.org/hofng/hofApp/domain/entity"

type UserAppInterface interface {
	CreateUser(user entity.User)
}

func (appHandler *app) CreateUser(user entity.User) {
	appHandler.repo.CreateUser(user)
}

package repository

import "bitbucket.org/hofng/hofApp/domain/entity"

type UserRepository interface {
	//	TODO
	CreateUser(user entity.User)
}

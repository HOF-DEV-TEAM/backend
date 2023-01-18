package application

import "bitbucket.org/hofng/hofApp/domain/repository"

type Applications interface {
	UserAppInterface
	ProductAppInterface
}
type app struct {
	repo repository.Repositories
}

func New(repo repository.Repositories) Applications {
	return &app{
		repo: repo,
	}
}

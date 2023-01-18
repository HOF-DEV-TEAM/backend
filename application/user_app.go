package application

type UserAppInterface interface {
	CreateUser()
}

func (appHandler *app) CreateUser() {
	appHandler.repo.CreateUser()
}

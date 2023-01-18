package application

type ProductAppInterface interface {
	CreateProduct()
}

func (appHandler *app) CreateProduct() {
	appHandler.repo.CreateProduct()
}

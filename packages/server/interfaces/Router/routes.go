package Router

import (
	httpSwagger "github.com/swaggo/http-swagger"

	_ "bitbucket.org/hofng/hofApp/docs"
	// "bitbucket.org/hofng/hofApp/domain/repository"
	"bitbucket.org/hofng/hofApp/interfaces"
	"github.com/go-chi/chi/v5"
)

func BuildRoutes(router *chi.Mux) {
	router.Handle("/swagger/*", httpSwagger.WrapHandler)

	//setup routes
	buildUserEndpoints(router)

}

func buildUserEndpoints(router *chi.Mux) {
	userRouter := chi.NewRouter()

	getUserHandler := interfaces.CreateGetUserHandler()
	userRouter.Get("/", getUserHandler)

	router.Mount("/user", userRouter)
}
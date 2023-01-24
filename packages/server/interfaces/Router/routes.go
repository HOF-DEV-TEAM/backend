package Router

import (
	"database/sql"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "bitbucket.org/hofng/hofApp/docs"
	"bitbucket.org/hofng/hofApp/pkg/user"

	"bitbucket.org/hofng/hofApp/interfaces"
	"github.com/go-chi/chi/v5"
)

func BuildRoutes(router *chi.Mux, logger *zap.Logger, db  *sql.DB) {
	router.Handle("/swagger/*", httpSwagger.WrapHandler)

	//setup routes
	userRepo := user.NewRepository(db, logger)
	userService := user.NewService(userRepo, logger)
	buildUserEndpoints(router, userService)

}

func buildUserEndpoints(router *chi.Mux, svc user.Service) {
	userRouter := chi.NewRouter()

	createUserHandler := interfaces.CreateGetUserHandler(svc)
	userRouter.Post("/", createUserHandler)

	router.Mount("/user", userRouter)
}

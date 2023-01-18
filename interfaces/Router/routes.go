package Router

import (
	_ "bitbucket.org/hofng/hofApp/docs"
	"bitbucket.org/hofng/hofApp/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

func Router(appPort, hostAddress string, httpHandler *interfaces.HTTPHandler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Mount("/hof", userEndpoints(httpHandler))
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(hostAddress+":"+appPort+"/swagger/doc.json"),
	))

	return router
}

func userEndpoints(httpHandler *interfaces.HTTPHandler) http.Handler {
	router := chi.NewRouter()
	router.Post("/user", httpHandler.CreateUser)
	return router
}

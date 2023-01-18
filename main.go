package main

import (
	"bitbucket.org/hofng/hofApp/application"
	"bitbucket.org/hofng/hofApp/infrastructure/library/logger"
	"bitbucket.org/hofng/hofApp/infrastructure/persistence"
	"bitbucket.org/hofng/hofApp/interfaces"
	"bitbucket.org/hofng/hofApp/interfaces/Router"
	"fmt"
	"log"
	"net/http"
)

// @title HOF BACKEND API
// @version 1.0
// @description This is the entire doc
// @termsOfService http://swagger.io/terms/
// @contact.name HOF Engineering
// @contact.email hof.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /hof
func main() {
	logger := logger.New()
	persistence, _, err := persistence.New("secrets.DatabaseURL", "secrets.DatabaseName", logger)
	if err != nil {
		//logger.Fatal("failed to open MongoDB", zap.Error(err))
	}
	applications := application.New(persistence)
	interfacesHandler := interfaces.New(applications)

	fmt.Println(interfacesHandler)

	//if err := client.Disconnect(context.Background()); err != nil {
	//	logger.Fatal("failed to disconnect from database", zap.Error(err))
	//}
	router := Router.Router("3000", "http://localhost", interfacesHandler)

	//helper.LogEvent("Info", fmt.Sprintf("Started UserServiceApplication on "+"http://localhost"+":"+"3000"+" in "+time.Since(time.Now()).String()))

	log.Fatal(http.ListenAndServe(":"+"3000", router))

}

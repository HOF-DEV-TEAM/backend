package main

import (
	"fmt"
	"os"

	"bitbucket.org/hofng/hofApp/application"
	"bitbucket.org/hofng/hofApp/infrastructure/library/logger"
)

//	@title			HOF BACKEND API
//	@version		1.0
//	@description	This is the entire doc
//	@termsOfService	http://swagger.io/terms/
//	@contact.name	HOF Engineering
//	@contact.email	hof.org
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			localhost:3000
//	@BasePath		/hof

func main() {
	logger := logger.New()

	app, err := application.New(logger)

	if err != nil {
		logger.Error(fmt.Sprintf("Fatal error creating application: %v", err))
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		logger.Error(fmt.Sprintf("Fatal error running application: %v", err))
		os.Exit(1)
	}
}

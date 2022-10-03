package rest

import (
	"github.com/labstack/echo/v4"
	"log"
	"os"
	"sofi/internal"
	"sofi/rest/routes"
)

var logger *log.Logger = log.New(os.Stdout, "api: ", log.LstdFlags|log.Lshortfile)
var local = false // TODO: Change to env variable.

func main() {
	if !local {
		err := internal.CreateRunners()
		if err != nil {
			logger.Fatalf("Error while trying to create runners: %v+", err)
		}

		err = internal.CreateUsers()
		if err != nil {
			logger.Fatalf("Error while trying to create users: %v+", err)
		}
	}

	err := internal.LoadLanguages()
	if err != nil {
		logger.Fatalf("Error while loading languages: %v+", err)
	}

	e := echo.New()
	e.GET("/languages", routes.Languages)
	e.POST("/execute", routes.Execute)

	e.Logger.Fatal(e.Start(":9090"))
}

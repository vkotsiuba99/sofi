package rest

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"sofi/pkg"
	"sofi/rest/routes"
	"strings"
)

var logger *log.Logger = log.New(os.Stdout, "api: ", log.LstdFlags|log.Lshortfile)

// loadEnv will load all the specified values from a specific string.
// It splits the string by comma and returns the origins.
func loadEnv(str string) []string {
	result := strings.Split(str, ",")
	for i := 0; i < len(result); i++ {
		result[i] = strings.TrimSpace(result[i])
	}

	return result
}

func main() {
	err := godotenv.Load("rest/local.env")
	if err != nil {
		logger.Fatalf("Error occurred while loading env file: %s", err)
	}

	origins := loadEnv(os.Getenv("ORIGINS"))
	activeLanguages := loadEnv(os.Getenv("LANGUAGES"))

	err = pkg.CreateRunners()
	if err != nil {
		logger.Fatalf("Error while trying to create runners: %v+", err)
	}

	err = pkg.CreateUsers()
	if err != nil {
		logger.Fatalf("Error while trying to create users: %v+", err)
	}

	err = pkg.LoadLanguages(activeLanguages)
	if err != nil {
		logger.Fatalf("Error while loading languages: %v+", err)
	}

	err = pkg.CreateBinaries()
	if err != nil {
		logger.Fatalf("Error while creating binaries: %v+", err)
	}

	e := echo.New()
	loggerGroup := e.Group("")

	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   middleware.NewRateLimiterMemoryStore(100),
		IdentifierExtractor: func(context echo.Context) (string, error) {
			id := context.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	// Only apply logging to the logger group.
	loggerGroup.Use(middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		// Check for empty request and response body (can occur when connecting with ws).
		if len(reqBody) == 0 && len(resBody) == 0 {
			return
		}

		pkg.Logger.Info(
			"request",
			zap.String("requestBody", string(reqBody)),
			zap.String("responseBody", string(resBody)),
		)
	}))

	rce := pkg.NewRceEngine()

	if _, err = pkg.NewLogger(pkg.ROTATE_WEEK); err != nil {
		logger.Fatalf("Error while creating logger: %v+", err)
	}
	logger.Println("Successfully created logger and connected to database.")
	defer pkg.CloseLogger()

	// Define REST endpoints.
	e.GET("/languages", routes.Languages)
	// Enables a websocket connection for piping the execution output.
	loggerGroup.GET("/execute", func(c echo.Context) error {
		return routes.ExecuteWs(c, rce)
	})
	// Post request for executing code without piping the output.
	loggerGroup.POST("/execute", func(c echo.Context) error {
		return routes.Execute(c, rce)
	})

	e.Logger.Fatal(e.Start(":9090"))
}

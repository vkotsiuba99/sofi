package routes

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
	"sofi/pkg"
)

type executeBody struct {
	Language string `json:"language" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

type executeResponse struct {
	Output string `json:"output"`
}

func Execute(c echo.Context) error {
	body := executeBody{}

	defer c.Request().Body.Close()

	err := json.NewDecoder(c.Request().Body).Decode(&body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	output, err := pkg.RunCode(body.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.JSON(http.StatusOK, executeResponse{
		Output: output.Result,
	})

	pkg.CleanUp(output.User, output.TempDirName)
	return nil
}

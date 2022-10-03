package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sofi/internal"
)

type languagesResponse struct {
	Languages []internal.Language `json:"languages"`
}

func Languages(c echo.Context) error {
	languages, err := internal.GetLanguages()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, languagesResponse{
		Languages: languages,
	})
}

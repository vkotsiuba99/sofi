package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sofi/pkg"
)

type languagesResponse struct {
	Languages []pkg.Language `json:"languages"`
}

func Languages(c echo.Context) error {
	languages, err := pkg.GetLanguages()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, languagesResponse{
		Languages: languages,
	})
}

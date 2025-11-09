package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/handlers"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/vmkteam/embedlog"
)

func NewRouter(logger embedlog.Logger, userHandler *handlers.UserHandler) *echo.Echo {
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", logger)
			return next(c)
		}
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/test", userHandler.GetUserById)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e
}

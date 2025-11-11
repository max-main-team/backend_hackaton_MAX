package http

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/handlers"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services/auth"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/vmkteam/embedlog"
)

func NewRouter(logger embedlog.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler, jwtService auth.JWTService, uniHandler *handlers.UniHandler, personsHandler *handlers.PersonalitiesHandler, facultiesHandler *handlers.FaculHandler) *echo.Echo {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${method} ${uri} ${status} ${latency_human} ` +
			`from=${remote_ip} ` +
			`user_agent="${user_agent}" ` +
			`error="${error}"` + "\n",
		Output: os.Stdout,
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://hackaton-max.vercel.app",
			"https://msokovykh.ru",
			"https://www.msokovykh.ru", // добавьте www
		},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", logger)
			return next(c)
		}
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// e.GET("/test", userHandler.GetUserById)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	protected := e.Group("")

	public := e.Group("")

	public.POST("/auth/login", authHandler.Login)
	public.POST("/auth/refresh", authHandler.Refresh)

	protected.Use(jwtService.JWTMiddleware())

	protected.GET("/auth/checkToken", authHandler.CheckToken)
	// protected.GET("/test", userHandler.GetUserById)

	admim := protected.Group("/admin")
	faculties := admim.Group("/faculties")
	faculties.POST("", facultiesHandler.GetFaculties)

	uni := protected.Group("/uni")
	uni.GET("/info", uniHandler.GetUniInfo)

	persons := protected.Group("/personalities")
	persons.POST("/access", personsHandler.RequestAccess)

	return e
}

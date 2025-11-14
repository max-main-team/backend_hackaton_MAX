package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/max-main-team/backend_hackaton_MAX/docs"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/handlers"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services/auth"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/vmkteam/embedlog"
)

func NewRouter(logger embedlog.Logger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	jwtService *auth.JWTService,
	uniHandler *handlers.UniHandler,
	personsHandler *handlers.PersonalitiesHandler,
	facultiesHandler *handlers.FaculHandler,
	subjectsHandler *handlers.SubjectHandler) *echo.Echo {
	e := echo.New()

	// e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	// 	Format: `[${time_rfc3339}] ${method} ${uri} ${status} ${latency_human} ` +
	// 		`from=${remote_ip} ` +
	// 		`user_agent="${user_agent}" ` +
	// 		`error="${error}"` + "\n",
	// 	Output: os.Stdout,
	// }))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{

		AllowOrigins: []string{
			"https://hackaton-max.vercel.app",
			"https://msokovykh.ru",
			"https://www.msokovykh.ru",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodPatch,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
			"X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           86400,
		ExposeHeaders:    []string{"Set-Cookie"},
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
		return c.String(http.StatusOK, "Костя лошара")
	})

	protected := e.Group("")

	public := e.Group("")

	public.POST("/auth/login", authHandler.Login)
	public.POST("/auth/refresh", authHandler.Refresh)

	protected.Use(jwtService.JWTMiddleware())

	users := protected.Group("/user")

	users.GET("/me", userHandler.GetUserInfo)

	protected.GET("/auth/checkToken", authHandler.CheckToken)
	// protected.GET("/test", userHandler.GetUserById)

	admin := protected.Group("/admin")
	faculties := admin.Group("/faculties")
	faculties.GET("", facultiesHandler.GetFaculties)
	faculties.POST("", facultiesHandler.CreateNewFaculty)

	uni := protected.Group("/universities")

	uni.GET("/info", uniHandler.GetUniInfo)

	uni.POST("/semesters", uniHandler.CreateNewSemesterPeriod)

	// get info about all universities
	uni.GET("/", uniHandler.GetAllUniversities)

	// personalities Admin
	persons := admin.Group("/personalities")
	persons.POST("/access", personsHandler.RequestAccess)
	persons.GET("/access", personsHandler.GetRequests)
	persons.POST("/access/accept", personsHandler.AcceptAccess)

	// personalities
	protected.GET("/personalities/universities", personsHandler.GetAllUniversitiesForPerson)
	protected.GET("/personalities/faculty", personsHandler.GetAllFacultiesForUniversity)
	protected.GET("/personalities/departments", personsHandler.GetAllDepartmentsForFaculty)
	protected.GET("/personalities/groups", personsHandler.GetAllGroupsForDepartment)
	protected.GET("/personalities/student", personsHandler.GetAllStudentForGtoup)
	protected.GET("/personalities/teachers", personsHandler.GetAllTeachersForUniversity)

	// subjects
	subjects := admin.Group("/subjects")
	subjects.POST("", subjectsHandler.Create)

	return e
}

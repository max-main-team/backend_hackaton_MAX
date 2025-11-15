package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	cfg "github.com/max-main-team/backend_hackaton_MAX/cfg"
	"github.com/max-main-team/backend_hackaton_MAX/internal/bot"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/handlers"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services/auth"
	"github.com/vmkteam/embedlog"
)

const (
	api_key_bot = "max_bot"
)

type App struct {
	sl      embedlog.Logger
	appName string
	cfg     cfg.Config
	db      *pgxpool.Pool
	echo    *echo.Echo
	bot     *bot.Bot

	jwtService       *auth.JWTService
	userHandler      *handlers.UserHandler
	authHandler      *handlers.AuthHandler
	uniHandler       *handlers.UniHandler
	personsHandler   *handlers.PersonalitiesHandler
	facultiesHandler *handlers.FaculHandler
	subjectsHandler  *handlers.SubjectHandler
	schedulesHandler *handlers.SchedulesHandler
}

func New(appName string, slogger embedlog.Logger, c cfg.Config, db *pgxpool.Pool) *App {
	a := &App{
		appName: appName,
		cfg:     c,
		db:      db,
		sl:      slogger,
	}
	a.initDependencies()
	a.echo = http.NewRouter(a.sl,
		a.userHandler,
		a.authHandler,
		a.jwtService,
		a.uniHandler,
		a.personsHandler,
		a.facultiesHandler,
		a.subjectsHandler,
		a.schedulesHandler)
	return a
}

type Dependencies struct {
	UserHandler *handlers.UserHandler
}

func (a *App) initDependencies() {

	// init repositories
	userRepo := repositories.NewUserRepository(a.db)
	refreshRepo, _ := repositories.NewPostgresRefreshTokenRepo(a.db)
	uniRepo := repositories.NewUniRepository(a.db)
	personsRepo := repositories.NewPersonalitiesRepo(a.db)
	faculRepo := repositories.NewFaculRepository(a.db)
	subjectsRepo := repositories.NewSubjectRepo(a.db)
	schedsRepo := repositories.NewScheduleRepo(a.db)

	// init services
	userService := services.NewUserService(userRepo)
	a.jwtService = auth.NewJWTService(a.cfg)
	uniService := services.NewUniService(uniRepo)
	faculService := services.NewFaculService(faculRepo)
	personService := services.NewPersonalitiesService(personsRepo)
	subjectsService := services.NewSubjectService(subjectsRepo)
	schedsService := services.NewSchedulesService(schedsRepo)

	// init handlers
	a.userHandler = handlers.NewUserHandler(userService, a.sl)
	a.authHandler = handlers.NewAuthHandler(
		a.jwtService,
		userRepo,
		refreshRepo,
		a.cfg.APIKeys[api_key_bot],
	)

	a.uniHandler = handlers.NewUniHandler(uniService, userService, a.sl)

	a.personsHandler = handlers.NewPersonalitiesHandler(personService, userService, a.sl)
	a.facultiesHandler = handlers.NewFaculHandler(faculService, userService, a.sl)
	a.subjectsHandler = handlers.NewSubjectHandler(subjectsService, userService, a.sl)
	a.schedulesHandler = handlers.NewSchedulesHandler(schedsService, userService, a.sl)

	if a.jwtService == nil {
		panic("jwt service is nil")
	}

	// init bot
	if botToken, ok := a.cfg.APIKeys[api_key_bot]; ok && botToken != "" {
		maxBot, err := bot.New(botToken, a.sl)
		if err != nil {
			a.sl.Errorf("Failed to create bot: %v", err)
		} else {
			a.bot = maxBot
			a.sl.Print(context.Background(), "Bot initialized successfully")
		}
	} else {
		a.sl.Print(context.Background(), "Bot token not found in config, bot will not be started")
	}
}
func (a *App) Run(ctx context.Context) error {
	// Запускаем бота в отдельной горутине, если он инициализирован
	if a.bot != nil {
		go func() {
			a.sl.Print(ctx, "Starting bot goroutine...")
			if err := a.bot.Start(ctx); err != nil {
				a.sl.Errorf("Bot stopped with error: %v", err)
			}
		}()
		a.sl.Print(ctx, "Bot started successfully")
	}

	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	a.sl.Print(ctx, "starting server", "addr", addr)
	return a.echo.Start(addr)
}

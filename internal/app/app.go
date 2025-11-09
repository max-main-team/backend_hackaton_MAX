package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	cfg "github.com/max-main-team/backend_hackaton_MAX/cfg"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/handlers"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type App struct {
	sl      embedlog.Logger
	appName string
	cfg     cfg.Config
	db      *pgxpool.Pool
	echo    *echo.Echo

	userHandler *handlers.UserHandler
}

func New(appName string, slogger embedlog.Logger, c cfg.Config, db *pgxpool.Pool) *App {
	a := &App{
		appName: appName,
		cfg:     c,
		db:      db,
		sl:      slogger,
	}
	a.initDependencies()

	a.echo = http.NewRouter(a.sl, a.userHandler)
	return a
}

type Dependencies struct {
	UserHandler *handlers.UserHandler
}

func (a *App) initDependencies() {

	// init repositories
	userRepo := repositories.NewUserRepository(a.db)

	// init services
	userService := services.NewUserService(userRepo)

	// init handlers
	a.userHandler = handlers.NewUserHandler(userService, a.sl)

}
func (a *App) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	a.sl.Print(ctx, "starting server", "addr", addr)
	return a.echo.Start(addr)
}

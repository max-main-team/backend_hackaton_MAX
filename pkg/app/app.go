package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	cfg "github.com/ssokov/backend_hackaton-MAX/cfg"
	"github.com/ssokov/backend_hackaton-MAX/pkg/http"
	"github.com/vmkteam/embedlog"
)

type App struct {
	sl      embedlog.Logger
	appName string
	cfg     cfg.Config
	db      *pgxpool.Pool
	echo    *echo.Echo
}

func New(appName string, slogger embedlog.Logger, c cfg.Config, db *pgxpool.Pool) *App {
	a := &App{
		appName: appName,
		cfg:     c,
		db:      db,
		sl:      slogger,
	}
	a.echo = http.NewRouter(a.sl)
	return a
}

func (a *App) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	a.sl.Print(ctx, "starting server", "addr", addr)
	return a.echo.Start(addr)
}

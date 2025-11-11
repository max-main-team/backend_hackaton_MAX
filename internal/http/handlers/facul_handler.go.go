package handlers

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type FaculHandler struct {
	faculService *services.FaculService
	logger       embedlog.Logger
}

func NewFaculHandler(service *services.FaculService, logger embedlog.Logger) *FaculHandler {
	return &FaculHandler{
		faculService: service,
		logger:       logger,
	}
}

func (u *FaculHandler) GetUniInfo(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "GetUniInfo called")
	return nil
}

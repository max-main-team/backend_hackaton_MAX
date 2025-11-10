package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type UserHandler struct {
	userService *services.UserService
	logger      embedlog.Logger
}

func NewUserHandler(service *services.UserService, logger embedlog.Logger) *UserHandler {
	return &UserHandler{
		userService: service,
		logger:      logger,
	}
}
func (h *UserHandler) GetUserById(c echo.Context) error {

	contextLogger := c.Get("logger").(embedlog.Logger)

	contextLogger.Print(context.Background(), "GetUserById called")

	user, _ := h.userService.GetUser(context.Background(), 1)

	return c.JSON(http.StatusOK, dto.UserResponse{
		ID: user.ID,
	})
}

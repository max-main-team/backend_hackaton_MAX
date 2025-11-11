package handlers

import (
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

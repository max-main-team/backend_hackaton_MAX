package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
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

func (u *UserHandler) GetUserInfo(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetUserInfo] GetUserInfo called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetUserInfo] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	userInfo, err := u.userService.GetUser(context.TODO(), currentUser.ID)

	if err != nil {
		log.Errorf("[GetUserInfo] Failed get user info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed get user info")
	}

	userRoles, err := u.userService.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[GetUserInfo] Failed find user roles: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed find user roles")
	}

	return c.JSON(http.StatusOK, dto.UserInfoResponse{
		UserRoles: userRoles.Roles,
		User: dto.User{
			ID:        int(userInfo.ID),
			FirstName: userInfo.FirstName,
			LastName:  PointerToString(userInfo.LastName),
			Username:  PointerToString(userInfo.UserName),
			PhotoURL:  PointerToString(userInfo.AvatarUrl),
		},
	})
}

func PointerToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

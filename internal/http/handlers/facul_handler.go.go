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

type FaculHandler struct {
	faculService *services.FaculService
	userService  *services.UserService
	logger       embedlog.Logger
}

func NewFaculHandler(service *services.FaculService, logger embedlog.Logger) *FaculHandler {
	return &FaculHandler{
		faculService: service,
		logger:       logger,
	}
}

func (f *FaculHandler) CreateNewFaculty(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[CreateNewFaculty] CreateNewFaculty called")

	var req dto.CreateNewFacultyRequest
	err := c.Bind(&req)
	if err != nil {
		log.Errorf("invalid request data. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid request data")
	}

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[CreateNewFaculty] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	roles, err := f.userService.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewFaculty] fail to get user roles. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user roles")
	}

	isAdmin := false
	for _, role := range roles.Roles {
		if role == "admim" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		log.Errorf("[CreateNewFaculty] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	err = f.faculService.CreateNewFaculty(context.TODO(), req.Name, currentUser.ID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "failed create new faculty")
	}
	return c.JSON(http.StatusOK, "faculty created successfully")
}

func (f *FaculHandler) GetFaculties(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetFaculties] GetUniInfo called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetFaculties] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	roles, err := f.userService.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[GetFaculties] fail to get user roles. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user roles")
	}

	isAdmin := false
	for _, role := range roles.Roles {
		if role == "admim" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		log.Errorf("[GetFaculties] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	faculties, err := f.faculService.GetInfoAboutUni(context.TODO(), currentUser.ID)

	if err != nil {
		log.Errorf("[GetFaculties] failed get faculties. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed get faculties")
	}

	return c.JSON(http.StatusOK, faculties)
}

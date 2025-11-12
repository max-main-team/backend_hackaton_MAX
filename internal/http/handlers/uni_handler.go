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

type UniHandler struct {
	uniService *services.UniService
	logger     embedlog.Logger
}

func NewUniHandler(service *services.UniService, logger embedlog.Logger) *UniHandler {
	return &UniHandler{
		uniService: service,
		logger:     logger,
	}
}

func (u *UniHandler) GetUniInfo(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "GetUniInfo called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	uniInfo, err := u.uniService.GetInfoAboutUni(context.TODO(), currentUser.ID)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed get info about uni")
	}

	return c.JSON(http.StatusOK, dto.UniInfoResponse{
		ID:        uniInfo.ID,
		Name:      uniInfo.Name,
		ShortName: uniInfo.ShortName,
		City:      uniInfo.City,
	})
}

func (u *UniHandler) GetAllUniversities(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "GetAllUniversities called")

	universities, err := u.uniService.GetAllUniversities(context.TODO())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed get all universities")
	}

	var response []dto.UniInfoResponse
	for _, uni := range universities {
		response = append(response, dto.UniInfoResponse{
			ID:          uni.ID,
			Name:        uni.Name,
			City:        uni.City,
			ShortName:   uni.ShortName,
			SiteUrl:     NewString(uni.SiteUrl),
			Description: NewString(uni.Description),
		})
	}

	return c.JSON(http.StatusOK, response)
}

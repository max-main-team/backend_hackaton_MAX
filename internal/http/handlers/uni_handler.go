package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type UniHandler struct {
	uniService  *services.UniService
	userService *services.UserService
	logger      embedlog.Logger
}

func NewUniHandler(service *services.UniService, logger embedlog.Logger) *UniHandler {
	return &UniHandler{
		uniService: service,
		logger:     logger,
	}
}

func (u *UniHandler) GetUniInfo(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetUniInfo] GetUniInfo called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetUniInfo] Authentication error")
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	uniInfo, err := u.uniService.GetInfoAboutUni(context.TODO(), currentUser.ID)

	if err != nil {
		log.Errorf("[GetUniInfo] Failed get info about uni. err: %v", err)
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

	log.Print(context.Background(), "[GetAllUniversities] GetAllUniversities called")

	universities, err := u.uniService.GetAllUniversities(context.TODO())

	if err != nil {
		log.Errorf("[GetAllUniversities] failed get all universities. err: %v", err)
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

// CreateNewNewSemesterPeriod godoc
// @Summary      Create new semester periods for university
// @Description  Create or replace semester periods for specific university. Admin role required. This operation will delete all existing semesters for the university and create new ones.
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        request  body   dto.CreateSemestersRequest  true  "Semester periods data"
// @Success      200   {object}  map[string]string  "status: semesters created successfully"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body or date format"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      403   {object}  echo.HTTPError  "Forbidden - user is not admin"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /universities/semesters [post]
// @Security     BearerAuth
func (u *UniHandler) CreateNewNewSemesterPeriod(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[CreateSemesters] CreateSemesters called")

	var req dto.CreateSemestersRequest

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CreateSemesters] Invalid request format: %v ", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}
	roles, err := u.userService.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[CreateSemesters] fail to get user roles. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user roles")
	}

	isAdmin := false
	for _, role := range roles.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		log.Errorf("[CreateSemesters] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	periods, err := ConvertDtoModel(req.Periods)
	if err != nil {
		log.Errorf("[CreateSemesters] failed convert string time -> time.Time. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed convert string time -> time.Time")
	}

	err = u.uniService.SetNewSemesterPeriod(context.TODO(), int64(req.ID), periods)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed create semesters")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "semesters created successfully"})
}

func ConvertDtoModel(dto []dto.SemesterPeriod) ([]models.SemesterPeriod, error) {
	out := make([]models.SemesterPeriod, 0, len(dto))
	for _, val := range dto {

		startDate, err := time.Parse(time.RFC3339, val.StartDate)
		if err != nil {
			return nil, err
		}

		endDate, err := time.Parse(time.RFC3339, val.EndDate)
		if err != nil {
			return nil, err
		}

		out = append(out, models.SemesterPeriod{
			StartDate: startDate,
			EndDate:   endDate,
		})
	}
	return out, nil
}

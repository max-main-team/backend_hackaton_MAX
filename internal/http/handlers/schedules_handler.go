package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/schedules"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type SchedulesHandler struct {
	schedulesServ *services.SchedulesService
	userServ      *services.UserService
	logger        embedlog.Logger
}

func NewSchedulesHandler(
	schedulesServ *services.SchedulesService,
	userServ *services.UserService,
	logger embedlog.Logger,
) *SchedulesHandler {
	return &SchedulesHandler{
		schedulesServ: schedulesServ,
		userServ:      userServ,
		logger:        logger,
	}
}

// CreateClass godoc
// @Summary create class (pair) slot
// @Tags schedules
// @Accept json
// @Produce json
// @Param request body schedules.CreateClassRequest true "Class info"
// @Success 200 {object} string "class_id"
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/classes [post]
func (h *SchedulesHandler) CreateClass(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[CreateClass] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	var req schedules.CreateClassRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		log.Errorf("[CreateClass] decode error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	id, err := h.schedulesServ.CreateClass(context.TODO(), schedules.CreateClassRequest{
		UniversityID: req.UniversityID,
		PairNumber:   req.PairNumber,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		log.Errorf("[CreateClass] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]int64{"id": id})
}

// DeleteClass godoc
// @Summary delete class
// @Tags schedules
// @Param class_id path int true "Class ID"
// @Success 200 {object} string "ok"
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/classes/{class_id} [delete]
func (h *SchedulesHandler) DeleteClass(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[DeleteClass] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	classIDStr := c.Param("class_id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		log.Errorf("[DeleteClass] parse class_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid class_id")
	}

	if err := h.schedulesServ.DeleteClass(context.TODO(), classID); err != nil {
		log.Errorf("[DeleteClass] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "ok")
}

// GetClassesByUniversity godoc
// @Summary get classes for university
// @Tags schedules
// @Produce json
// @Param university_id query int true "University ID"
// @Success 200 {array} schedules.ClassesResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/classes [get]
func (h *SchedulesHandler) GetClassesByUniversity(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetClassesByUniversity] called")

	universityIDStr := c.QueryParam("university_id")
	universityIS, err := strconv.ParseInt(universityIDStr, 10, 64)
	if err != nil {
		log.Errorf("[GetClassesByUniversity] parse university_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid university_id")
	}

	classes, err := h.schedulesServ.GetClassesByUniversity(context.TODO(), universityIS)
	if err != nil {
		log.Errorf("[GetClassesByUniversity] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, classes)
}

func (h *SchedulesHandler) requireAdmin(c echo.Context) (*models.User, error) {
	log := c.Get("logger").(embedlog.Logger)

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[requireAdmin] user not found in context")
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := h.userServ.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[requireAdmin] GetUserRolesByID error: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get roles")
	}

	hasAdmin := false
	for _, r := range roles.Roles {
		if r == "admin" {
			hasAdmin = true
			break
		}
	}

	if !hasAdmin {
		log.Errorf("[requireAdmin] user is not admin")
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	}

	return currentUser, nil
}

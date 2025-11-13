package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/subjects"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type SubjectHandler struct {
	subjectService *services.SubjectService
	userService    *services.UserService
	logger         embedlog.Logger
}

func NewSubjectHandler(subjectService *services.SubjectService, userService *services.UserService, logger embedlog.Logger) *SubjectHandler {
	return &SubjectHandler{
		subjectService: subjectService,
		userService:    userService,
		logger:         logger,
	}
}

// Create godoc
// @Summary create subject for university
// @Description handler that provide creation of subject for university
// @Tags subjects
// @Accept json
// @Produce json
// @Param request body subjects.CreateSubjectRequest true "Create request"
// @Success 200 {object} string "ok"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/subjects [post]
func (h *SubjectHandler) Create(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[Create] Create subject called")

	currentUser, ok := c.Get("user").(*models.User)

	if !ok {
		log.Errorf("[Create] User not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := h.userService.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[RequestAccess] GetUserRolesByID error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "user is not authenticated")
	}
	hasAdmin := slices.ContainsFunc(roles.Roles, func(s string) bool {
		return s == "admin "
	})
	if !hasAdmin {
		log.Errorf("[RequestAccess] GetUserRolesByID role admin not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	}

	var request subjects.CreateSubjectRequest
	if err = json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	err = h.subjectService.Create(context.TODO(), request)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create subject")
	}

	return c.JSON(http.StatusOK, "ok")
}

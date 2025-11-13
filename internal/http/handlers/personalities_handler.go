package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	personalities2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/http/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services"
	"github.com/vmkteam/embedlog"
)

type PersonalitiesHandler struct {
	personServ *services.PersonalitiesService
	userServ   *services.UserService
	logger     embedlog.Logger
}

func NewPersonalitiesHandler(personServ *services.PersonalitiesService, userServ *services.UserService, logger embedlog.Logger) *PersonalitiesHandler {
	return &PersonalitiesHandler{
		personServ: personServ,
		userServ:   userServ,
		logger:     logger,
	}
}

// RequestAccess godoc
// @Summary      Request access to join a university
// @Description  Current authenticated user sends a request to get a role in a university (student/teacher/administration).
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        request  body   personalities2.RequestAccessToUniversity  true  "Access request"
// @Success      200   {object}  string  "ok"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/personalities/access [post]
func (h *PersonalitiesHandler) RequestAccess(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[RequestAccess] RequestAccess called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[RequestAccess] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := h.userServ.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[RequestAccess] GetUserRolesByID error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "user is not authenticated")
	}
	hasAdmin := slices.ContainsFunc(roles.Roles, func(s string) bool {
		return s == "admin"
	})
	if !hasAdmin {
		log.Errorf("[RequestAccess] GetUserRolesByID role admin not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	}

	var request personalities2.RequestAccessToUniversity

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		log.Errorf("[RequestAccess] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.personServ.SendAccessToAddInUniversity(context.TODO(), int64(currentUser.ID), request)
	if err != nil {
		log.Errorf("[RequestAccess] failed to send access request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "ok")
}

// GetRequests godoc
// @Summary      get all requests access for administration of university
// @Description  Current authenticated user sends a request to get a access requests to be in university (student/teacher/administration).
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        limit  query   int  true  "limit of requests max(50), default(5)"
// @Param		offset 	query 	int 	true "offset default(0)"
// @Success      200   {object}  personalities2.AccessRequestResponse  "Requests for administration"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/personalities/access [get]
func (h *PersonalitiesHandler) GetRequests(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetRequests] GetRequests called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetRequests] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := h.userServ.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[RequestAccess] GetUserRolesByID error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "user is not authenticated")
	}
	hasAdmin := slices.ContainsFunc(roles.Roles, func(s string) bool {
		return s == "admin"
	})
	if !hasAdmin {
		log.Errorf("[RequestAccess] GetUserRolesByID role admin not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	}

	params := c.QueryParams()
	limit := params.Get("limit")
	offset := params.Get("offset")

	var limitInt, offsetInt int64
	if limit != "" {
		limitInt, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			log.Errorf("[GetRequests] failed to parse limit: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		limitInt = 5
	}

	if offset != "" {
		offsetInt, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			log.Errorf("[GetRequests] failed to parse offset: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		offsetInt = 0
	}

	response, err := h.personServ.GetAccessRequest(context.TODO(), currentUser.ID, limitInt, offsetInt)
	if err != nil {
		log.Errorf("[GetRequests] failed to get access request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// AcceptAccess godoc
// @Summary      Accept Request for adding in University
// @Description  Accept Request of user that want to be (student/teacher/administration), for student field university_department_id is required, course_group_id can be skipped. For administrations university_id is required, faculty_id can be skipped
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        request  body   personalities2.AcceptAccessRequest  true  "Access request"
// @Success      200   {object}  string  "ok"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/personalities/access/accept [post]
func (h *PersonalitiesHandler) AcceptAccess(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[AcceptRequest] AcceptRequest called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[AcceptRequest] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := h.userServ.GetUserRolesByID(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[RequestAccess] GetUserRolesByID error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "user is not authenticated")
	}
	hasAdmin := slices.ContainsFunc(roles.Roles, func(s string) bool {
		return s == "admin"
	})
	if !hasAdmin {
		log.Errorf("[RequestAccess] GetUserRolesByID role admin not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	}

	var request personalities2.AcceptAccessRequest
	if err = json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		log.Errorf("[AcceptRequest] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	switch request.UserType {
	case personalities.Student:
		if request.UniversityDepartmentID == nil {
			err = fmt.Errorf("university department id is required for student")
		}
	case personalities.Admin:
		if request.UniversityID == nil {
			err = fmt.Errorf("university id is required for admin")
		}
	}

	if err != nil {
		log.Errorf("[AcceptRequest] failed to validate request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.personServ.AcceptAccess(context.TODO(), request)
	if err != nil {
		log.Errorf("[AcceptRequest] failed to send access request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "ok")
}

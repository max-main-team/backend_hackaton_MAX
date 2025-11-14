package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
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

	// roles, err := h.userServ.GetUserRolesByID(context.TODO(), currentUser.ID)
	// if err != nil {
	// 	log.Errorf("[RequestAccess] GetUserRolesByID error: %v", err)
	// 	return echo.NewHTTPError(http.StatusInternalServerError, "user is not authenticated")
	// }
	// hasAdmin := slices.ContainsFunc(roles.Roles, func(s string) bool {
	// 	return s == "admin"
	// })
	// if !hasAdmin {
	// 	log.Errorf("[RequestAccess] GetUserRolesByID role admin not found")
	// 	return echo.NewHTTPError(http.StatusUnauthorized, "user is not admin")
	// }

	var request personalities2.RequestAccessToUniversity

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		log.Errorf("[RequestAccess] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.personServ.SendAccessToAddInUniversity(context.TODO(), int64(currentUser.ID), request)
	if err != nil {
		log.Errorf("[RequestAccess] failed to send access request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "ok")
}

// RejectRequestAccess godoc
// @Summary reject access request
// @Description decline access request for user
// @Tags personalities
// @Accept json
// @Produce json
// @Param request_id query int true "request_id"
// @Success      200   {object}  string  "ok"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/personalities/access [delete]
func (h *PersonalitiesHandler) RejectRequestAccess(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[RejectRequestAccess] RejectRequestAccess called")
	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[RejectRequestAccess] Authentication error. user not found in context")
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

	requestID := c.QueryParam("request_id")
	if requestID == "" {
		log.Errorf("[RejectRequestAccess] invalid request_id")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request_id")
	}

	requestIDInt, err := strconv.ParseInt(requestID, 10, 64)
	if err != nil {
		log.Errorf("[RejectRequestAccess] failed to parse request_id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request_id")
	}

	err = h.personServ.RejectRequest(context.TODO(), requestIDInt)
	if err != nil {
		log.Errorf("[RejectRequestAccess] failed to reject request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reject request")
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
	if response == nil {
		response = &personalities2.AccessRequestResponse{
			HasMore: false,
		}
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

// GetAllUniversitiesForPerson godoc
// @Summary      Get all universities for authenticated person
// @Description  Get all universities where the authenticated user has access (as admin or student)
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Success      200   {object}  []dto.UniInfoResponse  "Universities"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/universities [get]
func (h *PersonalitiesHandler) GetAllUniversitiesForPerson(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllUniversitiesFromPerson] GetAllUniversitiesFromPerson called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetAllUniversitiesFromPerson] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	universities, err := h.personServ.GetAllUniversitiesForPerson(context.TODO(), currentUser.ID)
	if err != nil {
		log.Errorf("[GetAllUniversitiesFromPerson] failed to get all universities for person: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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

// GetAllFacultiesForUniversity godoc
// @Summary      Get all faculties for university
// @Description  Get all faculties for university by university ID
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        university_id  query   int  true  "University ID"
// @Success      200   {object}  []dto.FacultyInfoResponse  "Faculties"
// @Failure      400   {object}  echo.HTTPError  "Invalid request parameter"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/faculty [get]
func (h *PersonalitiesHandler) GetAllFacultiesForUniversity(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllFacultiesForUniversity] GetAllFacultiesForUniversity called")

	universityID := c.QueryParam("university_id")
	universityIDInt, err := strconv.ParseInt(universityID, 10, 64)
	if err != nil {
		log.Errorf("[GetAllFacultiesForUniversity] failed to parse university id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	faculties, err := h.personServ.GetAllFacultiesForUniversity(context.TODO(), universityIDInt)
	if err != nil {
		log.Errorf("[GetAllFacultiesForUniversity] failed to get all faculties for university: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var response []dto.FacultyInfoResponse
	for _, faculty := range faculties {
		response = append(response, dto.FacultyInfoResponse{
			ID:             faculty.ID,
			Name:           faculty.Name,
			UniversityName: faculty.UniversityName,
		})
	}
	return c.JSON(http.StatusOK, response)

}

// GetAllDepartmentsForFaculty godoc
// @Summary      Get all departments for faculty
// @Description  Get all departments for faculty by faculty ID
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        faculty_id  query   int  true  "Faculty ID"
// @Success      200   {object}  []dto.DepartmentInfoResponse  "Departments"
// @Failure      400   {object}  echo.HTTPError  "Invalid request parameter"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/departments [get]
func (h *PersonalitiesHandler) GetAllDepartmentsForFaculty(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllDepartmentsForFaculty] GetAllDepartmentsForFaculty called")

	facultyID := c.QueryParam("faculty_id")
	facultyIDInt, err := strconv.ParseInt(facultyID, 10, 64)
	if err != nil {
		log.Errorf("[GetAllDepartmentsForFaculty] failed to parse faculty id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	departments, err := h.personServ.GetAllDepartmentsForFaculty(context.TODO(), facultyIDInt)
	if err != nil {
		log.Errorf("[GetAllDepartmentsForFaculty] failed to get all departments for faculty: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var response []dto.DepartmentInfoResponse
	for _, department := range departments {
		response = append(response, dto.DepartmentInfoResponse{
			ID:          int(department.ID),
			Name:        department.Name,
			FacultyName: department.FacultyName,
			Code:        department.Code,
		})
	}
	return c.JSON(http.StatusOK, response)
}

// GetAllGroupsForDepartment godoc
// @Summary      Get all groups for department
// @Description  Get all course groups for department by department ID
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        department_id  query   int  true  "Department ID"
// @Success      200   {object}  []dto.GroupInfoResponse  "Groups"
// @Failure      400   {object}  echo.HTTPError  "Invalid request parameter"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/groups [get]
func (h *PersonalitiesHandler) GetAllGroupsForDepartment(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllGroupsForDepartment] GetAllGroupsForDepartment called")

	departmentID := c.QueryParam("department_id")
	departmentIDInt, err := strconv.ParseInt(departmentID, 10, 64)
	if err != nil {
		log.Errorf("[GetAllGroupsForDepartment] failed to parse department id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	groups, err := h.personServ.GetAllGroupsForDepartment(context.TODO(), departmentIDInt)
	if err != nil {
		log.Errorf("[GetAllGroupsForDepartment] failed to get all groups for department: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var response []dto.GroupInfoResponse
	for _, group := range groups {
		response = append(response, dto.GroupInfoResponse{
			ID:             group.ID,
			Name:           group.Name,
			CourseID:       group.CourseID,
			DepartmentName: group.DepartmentName,
			Code:           group.Code,
		})
	}
	return c.JSON(http.StatusOK, response)
}

// GetAllStudentForGtoup godoc
// @Summary      Get all students for group
// @Description  Get all students for course group by group ID
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        group_id  query   int  true  "Course Group ID"
// @Success      200   {object}  []dto.User  "Students"
// @Failure      400   {object}  echo.HTTPError  "Invalid request parameter"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/student [get]
func (h *PersonalitiesHandler) GetAllStudentForGtoup(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllStudentForGtoup] GetAllStudentForGtoup called")

	_, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetAllStudentForGtoup] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	groupID := c.QueryParam("group_id")
	groupIDInt, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		log.Errorf("[GetAllStudentForGtoup] failed to parse group id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	students, err := h.personServ.GetAllStudentsForGroup(context.TODO(), groupIDInt)
	if err != nil {
		log.Errorf("[GetAllStudentForGtoup] failed to get all students for group: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var response []dto.User
	for _, student := range students {
		var lastName, userName, photoURL string
		if student.LastName != nil {
			lastName = *student.LastName
		}
		if student.UserName != nil {
			userName = *student.UserName
		}
		if student.AvatarUrl != nil {
			photoURL = *student.AvatarUrl
		}
		response = append(response, dto.User{
			ID:        int(student.ID),
			FirstName: student.FirstName,
			LastName:  lastName,
			Username:  userName,
			PhotoURL:  photoURL,
		})
	}
	return c.JSON(http.StatusOK, response)
}

// GetAllTeachersForUniversity godoc
// @Summary      Get all teachers for university
// @Description  Get all teachers for university by university ID
// @Tags         personalities
// @Accept       json
// @Produce      json
// @Param        university_id  query   int  true  "University ID"
// @Success      200   {object}  []dto.User  "Teachers"
// @Failure      400   {object}  echo.HTTPError  "Invalid request parameter"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /personalities/teachers [get]
func (h *PersonalitiesHandler) GetAllTeachersForUniversity(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetAllTeachersForUniversity] GetAllTeachersForUniversity called")

	_, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetAllTeachersForUniversity] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	universityID := c.QueryParam("university_id")
	universityIDInt, err := strconv.ParseInt(universityID, 10, 64)
	if err != nil {
		log.Errorf("[GetAllTeachersForUniversity] failed to parse university id: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	teachers, err := h.personServ.GetAllTeachersForUniversity(context.TODO(), universityIDInt)
	if err != nil {
		log.Errorf("[GetAllTeachersForUniversity] failed to get all teachers for university: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var response []dto.User
	for _, teacher := range teachers {
		var lastName, userName, photoURL string
		if teacher.LastName != nil {
			lastName = *teacher.LastName
		}
		if teacher.UserName != nil {
			userName = *teacher.UserName
		}
		if teacher.AvatarUrl != nil {
			photoURL = *teacher.AvatarUrl
		}
		response = append(response, dto.User{
			ID:        int(teacher.ID),
			FirstName: teacher.FirstName,
			LastName:  lastName,
			Username:  userName,
			PhotoURL:  photoURL,
		})
	}
	return c.JSON(http.StatusOK, response)
}

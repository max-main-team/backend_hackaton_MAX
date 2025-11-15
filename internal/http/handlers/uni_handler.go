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

func NewUniHandler(uniService *services.UniService, userService *services.UserService, logger embedlog.Logger) *UniHandler {
	return &UniHandler{
		userService: userService,
		uniService:  uniService,
		logger:      logger,
	}
}

// GetUniInfo godoc
// @Summary      Get university information for current user
// @Description  Get detailed information about the university associated with the authenticated user
// @Tags         universities
// @Accept       json
// @Produce      json
// @Success      200   {object}  dto.UniInfoResponse  "University information"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized - user not authenticated"
// @Failure      500   {object}  echo.HTTPError  "Internal server error - failed to get university info"
// @Router       /universities/info [get]
// @Security     BearerAuth
func (u *UniHandler) GetUniInfo(c echo.Context) error {
	ctx := c.Request().Context()

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetUniInfo] GetUniInfo called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetUniInfo] Authentication error")
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	uniInfo, err := u.uniService.GetInfoAboutUni(ctx, currentUser.ID)

	if err != nil {
		log.Errorf("[GetUniInfo] Failed get info about uni. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed get info about uni")
	}

	return c.JSON(http.StatusOK, dto.UniInfoResponse{
		ID:          uniInfo.ID,
		Name:        uniInfo.Name,
		ShortName:   uniInfo.ShortName,
		City:        uniInfo.City,
		SiteUrl:     NewString(uniInfo.SiteUrl),
		Description: NewString(uniInfo.Description),
		PhotoUrl:    NewString(uniInfo.PhotoUrl),
	})
}

// GetAllUniversities godoc
// @Summary      Get all universities
// @Description  Get a list of all universities with their detailed information
// @Tags         universities
// @Accept       json
// @Produce      json
// @Success      200   {array}   dto.UniInfoResponse  "List of universities"
// @Failure      500   {object}  echo.HTTPError  "Internal server error - failed to get universities"
// @Router       /universities/ [get]
// @Security     BearerAuth
func (u *UniHandler) GetAllUniversities(c echo.Context) error {
	ctx := c.Request().Context()

	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetAllUniversities] GetAllUniversities called")

	universities, err := u.uniService.GetAllUniversities(ctx)

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
			PhotoUrl:    NewString(uni.PhotoUrl),
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
func (u *UniHandler) CreateNewSemesterPeriod(c echo.Context) error {
	ctx := c.Request().Context()

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
	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
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

	err = u.uniService.SetNewSemesterPeriod(ctx, int64(req.ID), periods)
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

// CreateNewDepartment godoc
// @Summary      Create new department
// @Description  Create a new department and link it to a specific faculty and university. Creates entry in universities.departments and universities.university_departments. Admin role required.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body   dto.CreateDepartmentRequest  true  "Department data (department_name required, department_code and alias_name optional)"
// @Success      200   {object}  map[string]string  "status: department created successfully"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body or missing required fields"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      403   {object}  echo.HTTPError  "Forbidden - user is not admin"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/department [post]
// @Security     BearerAuth
func (u *UniHandler) CreateNewDepartment(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[CreateNewDepartment] CreateNewDepartment called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[CreateNewDepartment] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewDepartment] fail to get user roles. err: %v", err)
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
		log.Errorf("[CreateNewDepartment] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	var req dto.CreateDepartmentRequest

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CreateNewDepartment] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if req.DepartmentName == "" {
		log.Errorf("[CreateNewDepartment] department name is required")
		return echo.NewHTTPError(http.StatusBadRequest, "department name is required")
	}

	err = u.uniService.CreateNewDepartment(ctx, req.DepartmentName, req.DepartmentCode, req.AliasName, req.FacultyID, req.UniversityID)
	if err != nil {
		log.Errorf("[CreateNewDepartment] failed to create new department: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create new department")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "department created successfully"})
}

// CreateNewCourse godoc
// @Summary      Create new course
// @Description  Create a new course for a specific university department. Creates entry in universities.courses. Admin role required.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body   dto.CreateCourseRequest  true  "Course data (start_date, end_date, university_department_id required)"
// @Success      200   {object}  map[string]string  "status: course created successfully"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body or missing required fields"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      403   {object}  echo.HTTPError  "Forbidden - user is not admin"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/courses [post]
// @Security     BearerAuth
func (u *UniHandler) CreateNewCourse(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[CreateNewCourse] CreateNewCourse called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[CreateNewCourse] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewCourse] fail to get user roles. err: %v", err)
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
		log.Errorf("[CreateNewCourse] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	var req dto.CreateCourseRequest

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CreateNewCourse] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if req.StartDate == "" {
		log.Errorf("[CreateNewCourse] start date is required")
		return echo.NewHTTPError(http.StatusBadRequest, "start date is required")
	}

	if req.EndDate == "" {
		log.Errorf("[CreateNewCourse] end date is required")
		return echo.NewHTTPError(http.StatusBadRequest, "end date is required")
	}

	if req.UniversityDepartment <= 0 {
		log.Errorf("[CreateNewCourse] university department ID is required")
		return echo.NewHTTPError(http.StatusBadRequest, "university department ID is required")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		log.Errorf("[CreateNewCourse] invalid start date format: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		log.Errorf("[CreateNewCourse] invalid end date format: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end date format, use YYYY-MM-DD")
	}

	err = u.uniService.CreateNewCourse(ctx, startDate, endDate, req.UniversityDepartment)
	if err != nil {
		log.Errorf("[CreateNewCourse] failed to create new course: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create new course")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "course created successfully"})
}

// GetAllCourses godoc
// @Summary      Get all courses for university
// @Description  Get all courses for the admin's university. Admin role required.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {array}   dto.CourseInfoResponse  "List of courses"
// @Failure      401  {object}  echo.HTTPError          "Unauthorized user"
// @Failure      403  {object}  echo.HTTPError          "Forbidden - user is not admin"
// @Failure      500  {object}  echo.HTTPError          "Internal server error"
// @Router       /admin/courses [get]
// @Security     BearerAuth
func (u *UniHandler) GetAllCourses(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetAllCourses] GetAllCourses called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetAllCourses] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[GetAllCourses] fail to get user roles. err: %v", err)
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
		log.Errorf("[GetAllCourses] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	// Получаем university_id из администратора
	uniInfo, err := u.uniService.GetInfoAboutUni(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[GetAllCourses] failed to get university info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get university info")
	}

	courses, err := u.uniService.GetAllCoursesByUniversityID(ctx, int64(uniInfo.ID))
	if err != nil {
		log.Errorf("[GetAllCourses] failed to get courses: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get courses")
	}

	// Конвертируем в DTO
	var response []dto.CourseInfoResponse
	for _, course := range courses {
		response = append(response, dto.CourseInfoResponse{
			ID:                   course.ID,
			StartDate:            course.StartDate.Format("2006-01-02"),
			EndDate:              course.EndDate.Format("2006-01-02"),
			UniversityDepartment: course.UniversityDepartment,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// CreateNewGroup godoc
// @Summary      Create new course group
// @Description  Create a new course group for a specific course. Creates entry in groups.course_groups. Admin role required.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body   dto.CreateGroupRequest  true  "Group data (group_name and course_id required)"
// @Success      200   {object}  map[string]string  "status: group created successfully"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body or missing required fields"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      403   {object}  echo.HTTPError  "Forbidden - user is not admin"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /admin/groups [post]
// @Security     BearerAuth
func (u *UniHandler) CreateNewGroup(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[CreateNewGroup] CreateNewGroup called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[CreateNewGroup] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewGroup] fail to get user roles. err: %v", err)
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
		log.Errorf("[CreateNewGroup] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	var req dto.CreateGroupRequest

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CreateNewGroup] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if req.GroupName == "" {
		log.Errorf("[CreateNewGroup] group name is required")
		return echo.NewHTTPError(http.StatusBadRequest, "group name is required")
	}

	if req.CourseID <= 0 {
		log.Errorf("[CreateNewGroup] course ID is required")
		return echo.NewHTTPError(http.StatusBadRequest, "course ID is required")
	}

	err = u.uniService.CreateNewGroup(ctx, req.GroupName, req.CourseID)
	if err != nil {
		log.Errorf("[CreateNewGroup] failed to create new group: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create new group")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "group created successfully"})
}

// GetAllEvents godoc
// @Summary      Get all events for user's university
// @Description  Get all events for the university associated with the authenticated user
// @Tags         universities
// @Accept       json
// @Produce      json
// @Success      200   {array}   dto.EventResponse  "List of events"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized - user not authenticated"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /universities/events [get]
// @Security     BearerAuth
func (u *UniHandler) GetAllEvents(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[GetAllEvents] GetAllEvents called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetAllEvents] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	uniInfo, err := u.uniService.GetInfoAboutUni(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[GetAllEvents] failed to get university info. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get university info")
	}

	events, err := u.uniService.GetAllEventsByUniversityID(ctx, int64(uniInfo.ID))
	if err != nil {
		log.Errorf("[GetAllEvents] failed to get events. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get events")
	}

	var response []dto.EventResponse
	for _, event := range events {
		response = append(response, dto.EventResponse{
			ID:           event.ID,
			UniversityID: event.UniversityID,
			Title:        event.Title,
			Description:  event.Description,
			PhotoUrl:     event.PhotoUrl,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// CreateNewEvent godoc
// @Summary      Create new event
// @Description  Create a new event for the university. Admin role required.
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        request  body   dto.CreateEventRequest  true  "Event data"
// @Success      200   {object}  map[string]string  "status: event created successfully"
// @Failure      400   {object}  echo.HTTPError  "Invalid request body or missing required fields"
// @Failure      401   {object}  echo.HTTPError  "Unauthorized user"
// @Failure      403   {object}  echo.HTTPError  "Forbidden - user is not admin"
// @Failure      500   {object}  echo.HTTPError  "Internal server error"
// @Router       /universities/events [post]
// @Security     BearerAuth
func (u *UniHandler) CreateNewEvent(c echo.Context) error {
	ctx := c.Request().Context()
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[CreateNewEvent] CreateNewEvent called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[CreateNewEvent] Authentication error. user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	roles, err := u.userService.GetUserRolesByID(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewEvent] fail to get user roles. err: %v", err)
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
		log.Errorf("[CreateNewEvent] permission denied for user id %d", currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "permission denied. need role admin")
	}

	uniInfo, err := u.uniService.GetInfoAboutUni(ctx, currentUser.ID)
	if err != nil {
		log.Errorf("[CreateNewEvent] failed to get university info. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get university info")
	}

	var req dto.CreateEventRequest

	if err := c.Bind(&req); err != nil {
		log.Errorf("[CreateNewEvent] failed to decode request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if req.Title == "" {
		log.Errorf("[CreateNewEvent] title is required")
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	if req.Description == "" {
		log.Errorf("[CreateNewEvent] description is required")
		return echo.NewHTTPError(http.StatusBadRequest, "description is required")
	}

	if req.PhotoUrl == "" {
		log.Errorf("[CreateNewEvent] photo_url is required")
		return echo.NewHTTPError(http.StatusBadRequest, "photo_url is required")
	}

	event := models.Event{
		UniversityID: int64(uniInfo.ID),
		Title:        req.Title,
		Description:  req.Description,
		PhotoUrl:     req.PhotoUrl,
	}

	err = u.uniService.CreateNewEvent(ctx, event)
	if err != nil {
		log.Errorf("[CreateNewEvent] failed to create event: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create event")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "event created successfully"})
}

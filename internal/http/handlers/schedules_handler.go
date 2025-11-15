package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/schedules"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
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

	return c.JSON(http.StatusOK, id)
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

// CreateRoom godoc
// @Summary create room
// @Tags schedules
// @Accept json
// @Produce json
// @Param request body schedules.CreateRoomRequest true "Room info"
// @Success 200 {object} string "id"
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/rooms [post]
func (h *SchedulesHandler) CreateRoom(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[CreateRoom] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	var req schedules.CreateRoomRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		log.Errorf("[CreateRoom] decode error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	id, err := h.schedulesServ.CreateRoom(context.TODO(), req)
	if err != nil {
		log.Errorf("[CreateRoom] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, id)
}

// DeleteRoom godoc
// @Summary delete room
// @Tags schedules
// @Param room_id path int true "Room ID"
// @Success 200 {object} string "ok"
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/rooms/{room_id} [delete]
func (h *SchedulesHandler) DeleteRoom(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[DeleteRoom] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		log.Errorf("[DeleteRoom] parse room_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room_id")
	}

	if err := h.schedulesServ.DeleteRoom(context.TODO(), roomID); err != nil {
		log.Errorf("[DeleteRoom] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "ok")
}

// GetRoomsByUniversity godoc
// @Summary get rooms for university
// @Tags schedules
// @Produce json
// @Param university_id query int true "University ID"
// @Success 200 {array} schedules.RoomsResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /schedules/rooms [get]
func (h *SchedulesHandler) GetRoomsByUniversity(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetRoomsByUniversity] called")

	universityIDStr := c.QueryParam("university_id")
	universityID, err := strconv.ParseInt(universityIDStr, 10, 64)
	if err != nil {
		log.Errorf("[GetRoomsByUniversity] parse university_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid university_id")
	}

	rooms, err := h.schedulesServ.GetRoomsByUniversity(context.TODO(), universityID)
	if err != nil {
		log.Errorf("[GetRoomsByUniversity] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, rooms)
}

// CreateLesson godoc
// @Summary      Create lesson (group schedule entry)
// @Description  Создать занятие для учебной группы или элективной группы. Учёт конфликтов, лекций и интервалов.
// @Tags         schedules
// @Accept       json
// @Produce      json
// @Param        request  body      schedules.CreateLessonRequest  true  "Lesson info"
// @Success      200      {object}  string    "id"
// @Failure      400      {object}  echo.HTTPError       "Invalid request body"
// @Failure      401      {object}  echo.HTTPError       "Unauthorized user"
// @Failure      409      {object}  echo.HTTPError       "Schedule conflict"
// @Failure      500      {object}  echo.HTTPError       "Internal server error"
// @Router       /schedules/lessons [post]
func (h *SchedulesHandler) CreateLesson(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[CreateLesson] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	var req schedules.CreateLessonRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		log.Errorf("[CreateLesson] decode error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if (req.CourseGroupSubjectID == nil && req.ElectiveGroupSubjectID == nil) ||
		(req.CourseGroupSubjectID != nil && req.ElectiveGroupSubjectID != nil) {
		return echo.NewHTTPError(http.StatusBadRequest, "exactly one of course_group_subject_id or elective_group_subject_id must be set")
	}

	lessonID, err := h.schedulesServ.CreateLesson(context.Background(), req)
	if err != nil {
		if errors.Is(err, repositories.ErrScheduleConflict) {
			log.Errorf("[CreateLesson] schedule conflict: %v", err)
			return echo.NewHTTPError(http.StatusConflict, "schedule conflict (group/teacher/student/room)")
		}
		log.Errorf("[CreateLesson] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create lesson")
	}

	return c.JSON(http.StatusOK, lessonID)
}

// DeleteLesson godoc
// @Summary      Delete lesson
// @Tags         schedules
// @Produce      json
// @Param        lesson_id  path      int  true  "Lesson ID"
// @Success      200        {object}  string  "ok"
// @Failure      400        {object}  echo.HTTPError  "Invalid lesson_id"
// @Failure      401        {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500        {object}  echo.HTTPError  "Internal server error"
// @Router       /schedules/lessons/{lesson_id} [delete]
func (h *SchedulesHandler) DeleteLesson(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[DeleteLesson] called")

	_, err := h.requireAdmin(c)
	if err != nil {
		return err
	}

	lessonIDStr := c.Param("lesson_id")
	lessonID, err := strconv.ParseInt(lessonIDStr, 10, 64)
	if err != nil {
		log.Errorf("[DeleteLesson] parse lesson_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid lesson_id")
	}

	if err := h.schedulesServ.DeleteLesson(context.Background(), lessonID); err != nil {
		log.Errorf("[DeleteLesson] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete lesson")
	}

	return c.JSON(http.StatusOK, "ok")
}

// GetUserSchedule godoc
// @Summary      Get weekly schedule for user
// @Description  Возвращает расписание пользователя по max_user_id (и как студента, и как преподавателя).
// @Tags         schedules
// @Produce      json
// @Param        user_id  path      int  true  "MAX user id"
// @Success      200      {array}  	schedules.LessonsResponse
// @Failure      400      {object}  echo.HTTPError  "Invalid user_id"
// @Failure      401      {object}  echo.HTTPError  "Unauthorized user"
// @Failure      500      {object}  echo.HTTPError  "Internal server error"
// @Router       /schedules/users/{user_id} [get]
func (h *SchedulesHandler) GetUserSchedule(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)
	log.Print(context.Background(), "[GetUserSchedule] called")

	currentUser, ok := c.Get("user").(*models.User)
	if !ok {
		log.Errorf("[GetUserSchedule] user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user is not authenticated")
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Errorf("[GetUserSchedule] parse user_id error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// Разрешаем либо самому себе, либо админу.
	if currentUser.ID != userID {
		roles, err := h.userServ.GetUserRolesByID(context.Background(), currentUser.ID)
		if err != nil {
			log.Errorf("[GetUserSchedule] GetUserRolesByID error: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user roles")
		}

		isAdmin := false
		for _, r := range roles.Roles {
			if r == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			log.Errorf("[GetUserSchedule] user is not owner and not admin")
			return echo.NewHTTPError(http.StatusUnauthorized, "user is not allowed to view this schedule")
		}
	}

	schedule, err := h.schedulesServ.GetUserSchedule(context.Background(), userID)
	if err != nil {
		log.Errorf("[GetUserSchedule] service error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user schedule")
	}

	return c.JSON(http.StatusOK, schedule)
}

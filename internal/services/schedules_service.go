package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/schedules"
	schedules2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/schedules"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type SchedulesService struct {
	repo repositories.SchedulesRepository
}

func NewSchedulesService(repo repositories.SchedulesRepository) *SchedulesService {
	return &SchedulesService{repo: repo}
}

func (s *SchedulesService) CreateClass(ctx context.Context, request schedules.CreateClassRequest) (int64, error) {
	class := schedules2.Class{
		UniversityID: request.UniversityID,
		PairNumber:   request.PairNumber,
		StartTime:    request.StartTime,
		EndTime:      request.EndTime,
	}

	return s.repo.CreateClass(ctx, class)
}

func (s *SchedulesService) DeleteClass(ctx context.Context, classID int64) error {
	return s.repo.DeleteClass(ctx, classID)
}

func (s *SchedulesService) GetClassesByUniversity(ctx context.Context, universityID int64) ([]schedules.ClassesResponse, error) {
	classes, err := s.repo.GetClassesByUniversity(ctx, universityID)
	if err != nil {
		return []schedules.ClassesResponse{}, err
	}

	var classesResponse []schedules.ClassesResponse
	for _, class := range classes {
		classesResponse = append(classesResponse, schedules.ClassesResponse{
			ID:           class.ID,
			UniversityID: class.UniversityID,
			PairNumber:   class.PairNumber,
			StartTime:    class.StartTime,
			EndTime:      class.EndTime,
		})
	}

	return classesResponse, nil
}
func (s *SchedulesService) CreateRoom(ctx context.Context, request schedules.CreateRoomRequest) (int64, error) {
	room := schedules2.Room{
		UniversityID: request.UniversityID,
		Room:         request.Room,
	}
	return s.repo.CreateRoom(ctx, room)
}

func (s *SchedulesService) DeleteRoom(ctx context.Context, roomID int64) error {
	return s.repo.DeleteRoom(ctx, roomID)
}

func (s *SchedulesService) GetRoomsByUniversity(ctx context.Context, universityID int64) ([]schedules.RoomsResponse, error) {
	rooms, err := s.repo.GetRoomsByUniversity(ctx, universityID)
	if err != nil {
		return []schedules.RoomsResponse{}, err
	}

	var roomsResponse []schedules.RoomsResponse
	for _, room := range rooms {
		roomsResponse = append(roomsResponse, schedules.RoomsResponse{
			ID:           room.ID,
			UniversityID: room.UniversityID,
			Room:         room.Room,
		})
	}

	return roomsResponse, nil
}

func (s *SchedulesService) CreateLesson(ctx context.Context, req schedules.CreateLessonRequest) (int64, error) {
	r := schedules2.CreateLesson{
		CourseGroupSubjectID:   req.CourseGroupSubjectID,
		ElectiveGroupSubjectID: req.ElectiveGroupSubjectID,
		Day:                    schedules2.DayType(req.Day),
		ClassID:                req.ClassID,
		RoomID:                 req.RoomID,
		Interval:               schedules2.IntervalType(req.Interval),
	}
	return s.repo.CreateLesson(ctx, r)
}

func (s *SchedulesService) DeleteLesson(ctx context.Context, lessonID int64) error {
	return s.repo.DeleteLesson(ctx, lessonID)
}

func (s *SchedulesService) GetUserSchedule(ctx context.Context, userID int64) (schedules.LessonsResponse, error) {
	lessons, err := s.repo.GetUserSchedule(ctx, userID)
	if err != nil {
		return schedules.LessonsResponse{}, err
	}

	var lessonResponse schedules.LessonsResponse

	for _, lesson := range lessons {
		lessonResponse.Schedule = append(lessonResponse.Schedule, schedules.LessonItem{
			LessonID:         lesson.LessonID,
			Day:              lesson.Day,
			Interval:         lesson.Interval,
			PairNumber:       lesson.PairNumber,
			StartTime:        lesson.StartTime,
			EndTime:          lesson.EndTime,
			RoomID:           lesson.RoomID,
			Room:             lesson.Room,
			SubjectName:      lesson.SubjectName,
			SubjectType:      lesson.SubjectType,
			TeacherID:        lesson.TeacherID,
			TeacherFirstName: lesson.TeacherFirstName,
			TeacherLastName:  lesson.TeacherLastName,
		})
	}
	return lessonResponse, nil
}

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

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type UniService struct {
	uniRepo repositories.UniRepository
}

func NewUniService(repo repositories.UniRepository) *UniService {
	return &UniService{uniRepo: repo}
}

func (u *UniService) GetInfoAboutUni(ctx context.Context, id int64) (*models.UniversitiesData, error) {

	uniData, err := u.uniRepo.GetUniInfoByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uniData, nil
}

func (u *UniService) GetAllUniversities(ctx context.Context) ([]models.UniversitiesData, error) {

	universities, err := u.uniRepo.GetAllUniversities(ctx)
	if err != nil {
		return nil, err
	}

	return universities, nil
}

func (u *UniService) SetNewSemesterPeriod(ctx context.Context, uniID int64, periods []models.SemesterPeriod) error {

	err := ValidateSemesters(periods)
	if err != nil {
		return fmt.Errorf("failed create semesters. Invalid semesters periods. err :%w", err)
	}

	err = u.uniRepo.CreateSemestersForUniversity(ctx, uniID, periods)
	if err != nil {
		return fmt.Errorf("failed create semesters. err :%w", err)
	}

	return nil
}

func ValidateSemesters(periods []models.SemesterPeriod) error {
	if len(periods) == 0 {
		return fmt.Errorf("no periods provided")
	}

	for i, current := range periods {

		if current.StartDate.IsZero() || current.EndDate.IsZero() {
			return fmt.Errorf("semester %d: empty dates", i)
		}

		if current.StartDate.After(current.EndDate) {
			return fmt.Errorf("semester %d: start date after end date", i)
		}

		if current.EndDate.Sub(current.StartDate) < 24*time.Hour {
			return fmt.Errorf("semester %d: duration less than 1 day", i)
		}

		if i > 0 {
			prev := periods[i-1]
			if current.StartDate.Before(prev.EndDate) {
				return fmt.Errorf("semester %d overlaps with previous", i)
			}
		}
	}

	return nil
}

func (u *UniService) CreateNewDepartment(ctx context.Context, departmentName string, facultyID, universityID int64) error {
	if departmentName == "" {
		return fmt.Errorf("department name cannot be empty")
	}
	if facultyID <= 0 {
		return fmt.Errorf("invalid faculty ID")
	}
	if universityID <= 0 {
		return fmt.Errorf("invalid university ID")
	}

	err := u.uniRepo.CreateNewDepartment(ctx, departmentName, facultyID, universityID)
	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

func (u *UniService) CreateNewGroup(ctx context.Context, groupName string, departmentID, facultyID, universityID int64) error {
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	if departmentID <= 0 {
		return fmt.Errorf("invalid department ID")
	}
	if facultyID <= 0 {
		return fmt.Errorf("invalid faculty ID")
	}
	if universityID <= 0 {
		return fmt.Errorf("invalid university ID")
	}

	err := u.uniRepo.CreateNewGroup(ctx, groupName, departmentID, facultyID, universityID)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

func (u *UniService) CreateNewEvent(ctx context.Context, event models.Event) error {
	if event.Title == "" {
		return fmt.Errorf("event title cannot be empty")
	}
	if event.Description == "" {
		return fmt.Errorf("event description cannot be empty")
	}
	if event.PhotoUrl == "" {
		return fmt.Errorf("event photo URL cannot be empty")
	}
	if event.UniversityID <= 0 {
		return fmt.Errorf("invalid university ID")
	}

	err := u.uniRepo.CreateNewEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (u *UniService) GetAllEventsByUniversityID(ctx context.Context, universityID int64) ([]models.Event, error) {
	if universityID <= 0 {
		return nil, fmt.Errorf("invalid university ID")
	}

	events, err := u.uniRepo.GetAllEventsByUniversityID(ctx, universityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return events, nil
}

package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/personalities"
	personalities2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type PersonalitiesService struct {
	PersonsRepo repositories.PersonalitiesRepository
}

func NewPersonalitiesService(personsRepo repositories.PersonalitiesRepository) *PersonalitiesService {
	return &PersonalitiesService{
		PersonsRepo: personsRepo,
	}
}

func (s *PersonalitiesService) SendAccessToAddInUniversity(ctx context.Context, userID int64, request personalities.RequestAccessToUniversity) error {

	access := personalities2.UniversityAccess{
		UserType:     request.UserType,
		UniversityID: request.UniversityID,
		UserID:       userID,
	}

	err := s.PersonsRepo.RequestUniversityAccess(ctx, access)
	if err != nil {
		return err
	}
	return nil
}

func (s *PersonalitiesService) GetAccessRequest(ctx context.Context, userID, limit, offset int64) (*personalities.AccessRequestResponse, error) {
	accesses, err := s.PersonsRepo.GetAccessRequest(ctx, userID, limit+1, offset)
	if err != nil {
		return nil, err
	}

	var response personalities.AccessRequestResponse

	if int64(len(accesses.Requests)) > limit {
		response.HasMore = true
	}

	if int64(len(accesses.Requests)) < limit {
		limit = int64(len(accesses.Requests))
	}

	if len(accesses.Requests) == 0 {
		return &response, nil
	}

	response.Data = []struct {
		UserID    int64                   `json:"user_id"`
		UserType  personalities2.RoleType `json:"role"`
		FirstName string                  `json:"first_name"`
		LastName  *string                 `json:"last_name,omitempty"`
		Username  *string                 `json:"username,omitempty"`
	}(accesses.Requests[:limit])

	return &response, nil
}

func (s *PersonalitiesService) AcceptAccess(ctx context.Context, request personalities.AcceptAccessRequest) error {
	err := s.PersonsRepo.AddNewUser(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (s *PersonalitiesService) GetAllUniversitiesForPerson(ctx context.Context, userID int64) ([]models.UniversitiesData, error) {
	universities, err := s.PersonsRepo.GetAllUniversitiesForPerson(ctx, userID)
	if err != nil {
		return nil, err
	}
	return universities, nil
}

func (s *PersonalitiesService) GetAllFacultiesForUniversity(ctx context.Context, universityID int64) ([]models.Faculties, error) {
	faculties, err := s.PersonsRepo.GetAllFacultiesForUniversity(ctx, universityID)
	if err != nil {
		return nil, err
	}
	return faculties, nil
}

func (s *PersonalitiesService) GetAllDepartmentsForFaculty(ctx context.Context, facultyID int64) ([]models.Departments, error) {
	departments, err := s.PersonsRepo.GetAllDepartmentsForFaculty(ctx, facultyID)
	if err != nil {
		return nil, err
	}
	return departments, nil
}

func (s *PersonalitiesService) GetAllGroupsForDepartment(ctx context.Context, departmentID int64) ([]models.Groups, error) {
	groups, err := s.PersonsRepo.GetAllGroupsForDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *PersonalitiesService) GetAllStudentsForGroup(ctx context.Context, groupID int64) ([]models.User, error) {
	students, err := s.PersonsRepo.GetAllStudentsForGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (s *PersonalitiesService) GetAllTeachersForUniversity(ctx context.Context, universityID int64) ([]models.User, error) {
	teachers, err := s.PersonsRepo.GetAllTeachersForUniversity(ctx, universityID)
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

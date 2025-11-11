package services

import (
	"context"

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
		UserType:     personalities2.RoleType(request.UserType),
		UniversityID: request.UniversityID,
		UserID:       userID,
	}

	err := s.PersonsRepo.RequestUniversityAccess(ctx, access)
	if err != nil {
		return err
	}
	return nil
}

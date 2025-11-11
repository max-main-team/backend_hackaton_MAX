package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type FaculService struct {
	faculRepo repositories.FaculRepository
}

func NewFaculService(repo repositories.FaculRepository) *FaculService {
	return &FaculService{faculRepo: repo}
}

func (u *FaculService) GetInfoAboutUni(ctx context.Context, id int64) ([]models.Faculties, error) {

	faculties, err := u.faculRepo.GetFaculsByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return faculties, nil
}

package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type FaculService struct {
	faculRepo repositories.FaculRepository
}

func NewFaculService(uniRepository repositories.UniRepository) *FaculService {
	return &FaculService{faculRepo: uniRepository}
}

func (u *FaculService) GetInfoAboutUni(ctx context.Context, id int) (*models.UniversitiesData, error) {

	return nil, nil
}

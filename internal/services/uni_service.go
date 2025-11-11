package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type UniService struct {
	uniRepo repositories.UniRepository
}

func NewUniService(repo repositories.UniRepository) *UniService {
	return &UniService{uniRepo: repo}
}

func (u *UniService) GetInfoAboutUni(ctx context.Context, id int) (*models.UniversitiesData, error) {

	uniData, err := u.uniRepo.GetUniInfoByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uniData, nil
}

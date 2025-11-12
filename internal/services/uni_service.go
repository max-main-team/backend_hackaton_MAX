package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
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

func (u *UniService) CreateSemesters(ctx context.Context, id int64, periods []dto.SemesterPeriod) (*models.UniversitiesData, error) {
	// err := u.uniRepo.CreateSemestersForUniversity(ctx, id, periods)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}

package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models/http/subjects"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type SubjectService struct {
	subjectsRepo repositories.SubjectsRepository
}

func NewSubjectService(subjectsRepo repositories.SubjectsRepository) *SubjectService {
	return &SubjectService{
		subjectsRepo: subjectsRepo,
	}
}

func (s *SubjectService) Create(ctx context.Context, request subjects.CreateSubjectRequest) error {
	err := s.subjectsRepo.Create(ctx, request.Name, request.UniversityID)
	if err != nil {
		return err
	}
	return nil
}

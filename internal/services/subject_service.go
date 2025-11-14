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

func (s *SubjectService) Get(ctx context.Context, request subjects.GetSubjectsRequest, limit, offset int64) (*subjects.SubjectsResponse, error) {
	subs, err := s.subjectsRepo.Get(ctx, request.UniversityID, limit+1, offset)
	if err != nil {
		return nil, err
	}

	var response subjects.SubjectsResponse

	if int64(len(subs.Data)) > limit {
		response.HasMore = true
	}

	// if int64(len(subs.Data)) < limit {
	// 	limit = int64(len(subs.Data))
	// }

	if len(subs.Data) == 0 {
		return &response, nil
	}

	for _, sub := range subs.Data {
		response.Data = append(response.Data, struct {
			Name string `json:"name"`
			ID   int64  `json:"id"`
		}{Name: sub.Name, ID: sub.ID})
	}

	return &response, nil
}

func (s *SubjectService) Delete(ctx context.Context, subjectID int64) error {
	err := s.subjectsRepo.Delete(ctx, subjectID)
	return err
}

package services

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (u *UserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	return u.userRepo.GetUserByID(ctx, id)
}

func (u *UserService) GetUserRolesByID(ctx context.Context, id int) (*models.UserRoles, error) {

	roles, err := u.userRepo.GetUserRolesByID(ctx, id)
	if roles != nil {
		return nil, err
	}
	return roles, nil
}

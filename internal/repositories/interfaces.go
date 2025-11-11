package repositories

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserRolesByID(ctx context.Context, id int) (*models.UserRoles, error)
}

type RefreshTokenRepository interface {
	Save(token *models.RefreshToken) error
	Find(tokenString string) (*models.RefreshToken, error)
	Delete(tokenString string) error
	DeleteByUser(userID int) error
}

type UniRepository interface {
	GetUniInfoByUserID(ctx context.Context, id int) (*models.UniversitiesData, error)
}

type FaculRepository interface{}

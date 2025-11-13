package repositories

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	personalities2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/http/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserRolesByID(ctx context.Context, id int64) (*models.UserRoles, error)
	CreateNewUser(ctx context.Context, user *models.User) error
}

type RefreshTokenRepository interface {
	Save(token *models.RefreshToken) error
	Find(tokenString string) (*models.RefreshToken, error)
	Delete(tokenString string) error
	DeleteByUser(userID int) error
}

type UniRepository interface {
	GetAllUniversities(ctx context.Context) ([]models.UniversitiesData, error)

	CreateSemestersForUniversity(ctx context.Context, uniID int64, periods []models.SemesterPeriod) error

	GetUniInfoByUserID(ctx context.Context, id int64) (*models.UniversitiesData, error)
}

type PersonalitiesRepository interface {
	RequestUniversityAccess(ctx context.Context, uniAccess personalities.UniversityAccess) error
	GetAccessRequest(ctx context.Context, userID, limit, offset int64) (personalities.AccessRequests, error)
	AddNewUser(ctx context.Context, request personalities2.AcceptAccessRequest) error
}

type FaculRepository interface {
	GetFaculsByUserID(ctx context.Context, id int64) ([]models.Faculties, error)
}

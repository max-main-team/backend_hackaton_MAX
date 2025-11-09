package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (u *userRepository) GetUserById(ctx context.Context, id int) (*models.User, error) {
	return &models.User{
		ID:       1,
		Username: "testuser",
	}, nil
}

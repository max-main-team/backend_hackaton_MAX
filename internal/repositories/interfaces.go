package repositories

import (
	"context"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id int) (*models.User, error)
}

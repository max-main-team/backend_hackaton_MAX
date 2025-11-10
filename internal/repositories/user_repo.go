package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (u *userRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, first_name, last_name, username, is_bot, last_activity_time, description, avatar_url, full_avatar_url
			  FROM users.max_users_data
			  WHERE id = $1`
	err := u.pool.QueryRow(ctx, query, id).Scan(user.ID,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
		&user.IsBot,
		&user.LastActivityTime,
		&user.Description,
		&user.AvatarUrl,
		&user.FullAvatarUrl)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

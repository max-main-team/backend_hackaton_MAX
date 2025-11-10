package repositories

import (
	"context"
	"fmt"

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

func (u *userRepository) GetUserRolesByID(ctx context.Context, id int) (*models.UserRoles, error) {
	var roles []string
	query :=
		`
	SELECT 'admin' as role 
	FROM personalities.administrations
	WHERE max_user_id = $1
	UNION 
	SELECT 'teacher' as role 
	FROM personalities.teachers 
	WHERE max_user_id = $1
	UNION
	SELECT 'student' as role 
	FROM personalities.students 
	WHERE max_user_id = $1
	`

	rows, err := u.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed GetUserRolesByID from db. err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("failed GetUserRolesByID from db in scan. err: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed GetUserRolesByID during iteration. err: %w", err)
	}

	rolesCopy := make([]string, len(roles))
	copy(rolesCopy, roles)

	userRoles := models.UserRoles{
		Roles: rolesCopy,
	}

	return &userRoles, nil
}

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

func (u *userRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, first_name, last_name, username, is_bot, last_activity, description, avatar_url, full_avatar_url
			  FROM users.max_users_data
			  WHERE id = $1`
	err := u.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
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

func (u *userRepository) CreateNewUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users.max_users_data (id, first_name, last_name, username, is_bot, last_activity, description, avatar_url, full_avatar_url)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := u.pool.Exec(ctx, query,
		user.ID,
		user.FirstName,
		toNullString(user.LastName),
		toNullString(user.UserName),
		user.IsBot,
		user.LastActivityTime,
		toNullString(user.Description),
		toNullString(user.AvatarUrl),
		toNullString(user.FullAvatarUrl))

	if err != nil {
		return fmt.Errorf("failed to create new user: %w", err)
	}

	return nil
}

func (u *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users.max_users_data 
		SET first_name = $2,
		    last_name = $3,
		    username = $4,
		    avatar_url = $5,
		    full_avatar_url = $6
		WHERE id = $1
	`

	_, err := u.pool.Exec(ctx, query,
		user.ID,
		user.FirstName,
		toNullString(user.LastName),
		toNullString(user.UserName),
		toNullString(user.AvatarUrl),
		toNullString(user.FullAvatarUrl))

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (u *userRepository) GetUserRolesByID(ctx context.Context, id int64) (*models.UserRoles, error) {
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

func toNullString(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

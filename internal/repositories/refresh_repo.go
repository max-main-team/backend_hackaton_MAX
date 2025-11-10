package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

var ErrRefreshNotFound = fmt.Errorf("refresh token not found")

type pgRefreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRefreshTokenRepo(pool *pgxpool.Pool) (RefreshTokenRepository, error) {
	return &pgRefreshTokenRepo{pool: pool}, nil
}

func (r *pgRefreshTokenRepo) Save(rt *models.RefreshToken) error {
	ctx := context.Background()
	const q = `
        INSERT INTO users.refresh_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
	return r.pool.QueryRow(ctx, q, rt.UserID, rt.Token, rt.ExpiresAt).
		Scan(&rt.ID, &rt.CreatedAt)
}

func (r *pgRefreshTokenRepo) Find(token string) (*models.RefreshToken, error) {
	ctx := context.Background()
	const q = `
        SELECT id, user_id, token, expires_at, created_at
        FROM users.refresh_tokens
        WHERE token = $1
    `
	var rt models.RefreshToken
	err := r.pool.QueryRow(ctx, q, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrRefreshNotFound
		}
		return nil, fmt.Errorf("Find refresh failed: %w", err)
	}
	return &rt, nil
}

func (r *pgRefreshTokenRepo) Delete(token string) error {
	ctx := context.Background()
	const q = `DELETE FROM users.refresh_tokens WHERE token = $1`
	cmd, err := r.pool.Exec(ctx, q, token)
	if err != nil {
		return fmt.Errorf("Delete refresh failed: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrRefreshNotFound
	}
	return nil
}

func (r *pgRefreshTokenRepo) DeleteByUser(userID int) error {
	ctx := context.Background()
	const q = `DELETE FROM users.refresh_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("DeleteByUser failed: %w", err)
	}
	return nil
}

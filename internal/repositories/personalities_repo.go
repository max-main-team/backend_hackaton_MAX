package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"
)

type PersonalitiesRepo struct {
	pool *pgxpool.Pool
}

func NewPersonalitiesRepo(pool *pgxpool.Pool) *PersonalitiesRepo {
	return &PersonalitiesRepo{pool: pool}
}

func (r *PersonalitiesRepo) RequestUniversityAccess(ctx context.Context, uniAccess personalities.UniversityAccess) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	const qSendAccess = `
		INSERT INTO users.persons_adds (
			from_max_user_id,
			to_max_user_id,
			role_type
		) 
		SELECT 
		    $1,
		    pa.id,
		    $2
		FROM personalities.administrations pa
		WHERE pa.university_id = $3
		ON CONFLICT DO NOTHING;
	`

	_, err = tx.Exec(ctx, qSendAccess, uniAccess.UserID, uniAccess.UserType, uniAccess.UniversityID)
	if err != nil {
		return err
	}

	return nil
}

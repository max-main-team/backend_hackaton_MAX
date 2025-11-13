package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubjectRepo struct {
	pool *pgxpool.Pool
}

func NewSubjectRepo(pool *pgxpool.Pool) *SubjectRepo {
	return &SubjectRepo{pool: pool}
}

func (r *SubjectRepo) Create(ctx context.Context, name string, uniID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
		err = tx.Commit(ctx)
	}()

	const qInsertSubject = `
		INSERT INTO subjects.university_subjects(university_id, name) VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, qInsertSubject, uniID, name)
	if err != nil {
		return err
	}

	return nil
}

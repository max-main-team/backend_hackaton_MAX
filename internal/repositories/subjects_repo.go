package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/subjects"
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

func (r *SubjectRepo) Get(ctx context.Context, uniID, limit, offset int64) (*subjects.Subjects, error) {
	const qGetSubjects = `
		SELECT 
			s.id,
			s.name
		FROM subjects.university_subjects as s
		WHERE s.university_id = $1
	`

	rows, err := r.pool.Query(ctx, qGetSubjects, uniID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs subjects.Subjects

	for rows.Next() {
		var s subjects.Subject
		err = rows.Scan(&s.ID, &s.Name)
		if err != nil {
			return nil, err
		}

		subs.Data = append(subs.Data, s)
	}

	return &subs, nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type faculRepository struct {
	pool *pgxpool.Pool
}

func NewFaculRepository(pool *pgxpool.Pool) FaculRepository {
	return &faculRepository{pool: pool}
}

func (u *faculRepository) GetFaculsByUserID(ctx context.Context, id int64) ([]models.Faculties, error) {

	var faculties []models.Faculties
	query := `
        SELECT uf.id, uf.name, uud.name

        FROM universities.faculties AS uf

        JOIN universities.universities_data AS uud

		ON uf.university_id = uud.id

        WHERE uud.id = (
            SELECT university_id 
            FROM personalities.administrations
            WHERE max_user_id = $1
        )
    `
	rows, err := u.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed GetFaculsByUserID from db. err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var facul models.Faculties
		if err := rows.Scan(&facul.ID, &facul.Name, &facul.UniversityName); err != nil {
			return nil, fmt.Errorf("failed GetFaculsByUserID from db in scan. err: %w", err)
		}
		faculties = append(faculties, facul)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed GetFaculsByUserID during iteration. err: %w", err)
	}

	return faculties, nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type uniRepository struct {
	pool *pgxpool.Pool
}

func NewUniRepository(pool *pgxpool.Pool) UniRepository {
	return &uniRepository{pool: pool}
}

func (u *uniRepository) GetUniInfoByUserID(ctx context.Context, id int) (*models.UniversitiesData, error) {
	var uniData models.UniversitiesData

	query := `
        SELECT uc.name, uud.name, uud.short_name
        FROM universities.universities_data AS uud
        JOIN universities.cities AS uc ON uud.city_id = uc.id
        WHERE uud.id = (
            SELECT university_id 
            FROM personalities.administrations
            WHERE max_user_id = $1
        )
    `

	err := u.pool.QueryRow(ctx, query, id).Scan(&uniData.City, &uniData.Name, &uniData.ShortName)

	if err != nil {
		return nil, fmt.Errorf("failed get info about uni. err: %v", err)
	}

	return &uniData, nil
}

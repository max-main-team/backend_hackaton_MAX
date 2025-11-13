package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

type uniRepository struct {
	pool *pgxpool.Pool
}

func NewUniRepository(pool *pgxpool.Pool) UniRepository {
	return &uniRepository{pool: pool}
}

func (u *uniRepository) GetUniInfoByUserID(ctx context.Context, id int64) (*models.UniversitiesData, error) {
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

func (u *uniRepository) GetAllUniversities(ctx context.Context) ([]models.UniversitiesData, error) {
	var universities []models.UniversitiesData

	query := `
        SELECT uud.id, uud.name, uc.name, uud.short_name, uud.site_url, uud.description
        FROM universities.universities_data AS uud
        JOIN universities.cities AS uc
		ON uud.city_id = uc.id
	`
	rows, err := u.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all universities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var uni models.UniversitiesData
		err := rows.Scan(&uni.ID, &uni.Name, &uni.City, &uni.ShortName, &uni.SiteUrl, &uni.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan university row: %w", err)
		}
		universities = append(universities, uni)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return universities, nil
}

func (u *uniRepository) CreateSemestersForUniversity(ctx context.Context, userID int64) error {
	log.Println("ssdaa")
	return nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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
        SELECT uc.name, uud.name, uud.short_name, uud.site_url, uud.description, uud.photo_url
        FROM universities.universities_data AS uud
        JOIN universities.cities AS uc ON uud.city_id = uc.id
        WHERE uud.id = (
            SELECT university_id 
            FROM personalities.administrations
            WHERE max_user_id = $1
        )
    `

	err := u.pool.QueryRow(ctx, query, id).Scan(&uniData.City, &uniData.Name, &uniData.ShortName, &uniData.SiteUrl, &uniData.Description, &uniData.PhotoUrl)

	if err != nil {
		return nil, fmt.Errorf("failed get info about uni. err: %v", err)
	}

	return &uniData, nil
}

func (u *uniRepository) GetAllUniversities(ctx context.Context) ([]models.UniversitiesData, error) {
	var universities []models.UniversitiesData

	query := `
        SELECT uud.id, uud.name, uc.name, uud.short_name, uud.site_url, uud.description, uud.photo_url
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
		err := rows.Scan(&uni.ID, &uni.Name, &uni.City, &uni.ShortName, &uni.SiteUrl, &uni.Description, &uni.PhotoUrl)
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

func (u *uniRepository) CreateSemestersForUniversity(ctx context.Context, uniID int64, periods []models.SemesterPeriod) error {

	deleteQuery :=
		`
	DELETE FROM universities.semesters
	WHERE university_id = $1

	`
	insertQuery :=
		`
	INSERT INTO universities.semesters (start_date,end_date,university_id)
	VALUES ($1,$2,$3)
	`

	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			_ = rollbackErr
		}
	}()

	_, err = tx.Exec(ctx, deleteQuery, uniID)
	if err != nil {
		return fmt.Errorf("failed delete semesters from db. err: %w", err)
	}

	for _, val := range periods {
		_, err := tx.Exec(ctx, insertQuery, val.StartDate, val.EndDate, uniID)
		if err != nil {
			return fmt.Errorf("failed to insert to db. err: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction. err: %w", err)
	}
	return nil
}

func (u *uniRepository) CreateNewDepartment(ctx context.Context, departmentName string, facultyID, universityID int64) error {
	query := `
		INSERT INTO universities.university_departments (name, faculty_id, university_id)
		VALUES ($1, $2, $3)
	`

	_, err := u.pool.Exec(ctx, query, departmentName, facultyID, universityID)
	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

func (u *uniRepository) CreateNewGroup(ctx context.Context, groupName string, departmentID, facultyID, universityID int64) error {
	query := `
		INSERT INTO universities.course_groups (name, university_department_id, faculty_id, university_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := u.pool.Exec(ctx, query, groupName, departmentID, facultyID, universityID)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

func (u *uniRepository) CreateNewEvent(ctx context.Context, event models.Event) error {
	query := `
		INSERT INTO universities.events (university_id, title, description, photo_url)
		VALUES ($1, $2, $3, $4)
	`

	_, err := u.pool.Exec(ctx, query, event.UniversityID, event.Title, event.Description, event.PhotoUrl)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (u *uniRepository) GetAllEventsByUniversityID(ctx context.Context, universityID int64) ([]models.Event, error) {
	var events []models.Event

	query := `
		SELECT id, university_id, title, description, photo_url
		FROM universities.events
		WHERE university_id = $1
		ORDER BY id DESC
	`

	rows, err := u.pool.Query(ctx, query, universityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.Event
		err := rows.Scan(&event.ID, &event.UniversityID, &event.Title, &event.Description, &event.PhotoUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

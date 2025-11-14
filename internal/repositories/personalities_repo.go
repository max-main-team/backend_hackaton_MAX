package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	personalities2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/http/personalities"
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
			to_administration_id,
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

func (r *PersonalitiesRepo) GetAccessRequest(ctx context.Context, userID, limit, offset int64) (personalities.AccessRequests, error) {
	const qGetAccessByUser = `
		SELECT
			u.from_max_user_id as user_id,
			u.role_type as role,
			mu.first_name as first_name,
			mu.last_name as last_name,
			mu.username as username,
		FROM users.persons_adds as u
		JOIN users.max_users_data mu on mu.id = u.from_max_user_id
		WHERE u.to_administration_id = 
		      (select a.id from personalities.administrations a where a.max_user_id = $1)
		ORDER BY user_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, qGetAccessByUser, userID, limit, offset)
	if err != nil {
		return personalities.AccessRequests{}, err
	}
	defer rows.Close()

	var result personalities.AccessRequests
	for rows.Next() {
		var userID int64
		var roleType personalities.RoleType
		var (
			firstName string
			lastName,
			username *string
		)
		if err := rows.Scan(&userID, &roleType, &firstName, &lastName, &username); err != nil {
			return personalities.AccessRequests{}, err
		}
		result.Requests = append(result.Requests, struct {
			UserID    int64
			UserType  personalities.RoleType
			FirstName string
			LastName  *string
			Username  *string
		}{userID, roleType, firstName, lastName, username})
	}

	return result, nil
}

func (r *PersonalitiesRepo) AddNewUser(ctx context.Context, request personalities2.AcceptAccessRequest) error {
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

	var qInsertUser string

	switch request.UserType {
	case personalities.Student:
		qInsertUser = fmt.Sprintf(`
			INSERT INTO users.students (
			                            max_user_id,
			                            university_department_id,
			                            course_group_id
			) VALUES (%d, %d, %d)
		`, request.UserID, request.UniversityDepartmentID, request.CourseGroupID)
	case personalities.Teacher:
		qInsertUser = fmt.Sprintf(`
			INSERT INTO users.teachers (
			                            max_user_id
			) VALUES (%d)
	`, request.UserID)
	case personalities.Admin:
		qInsertUser = fmt.Sprintf(`
		INSERT INTO users.administrations (
		                                   max_user_id,
		                                   university_id,
		                                	faculty_id
		) VALUES (%d, %d, %d)
`, request.UserID, request.UniversityID, request.FacultyID)
	}

	_, err = tx.Exec(ctx, qInsertUser)
	if err != nil {
		return err
	}

	return nil
}

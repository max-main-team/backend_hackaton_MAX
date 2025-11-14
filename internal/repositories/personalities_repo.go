package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
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
			mu.username as username
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
			INSERT INTO personalities.students (
			                            max_user_id,
			                            university_department_id,
			                            course_group_id
			) VALUES (%d, %d, %d)
		`, request.UserID, request.UniversityDepartmentID, request.CourseGroupID)
	case personalities.Teacher:
		qInsertUser = fmt.Sprintf(`
			INSERT INTO personalities.teachers (
			                            max_user_id
			) VALUES (%d)
	`, request.UserID)
	case personalities.Admin:
		qInsertUser = fmt.Sprintf(`
		INSERT INTO personalities.administrations (
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

func (r *PersonalitiesRepo) GetAllUniversitiesForPerson(ctx context.Context, userID int64) ([]models.UniversitiesData, error) {
	const qGetAllUniversitiesForPerson = `
		SELECT DISTINCT
			uud.id,
			uud.name,
			c.name as city,
			uud.short_name,
			uud.site_url,
			uud.description
		FROM universities.universities_data AS uud
		LEFT JOIN universities.cities AS c ON uud.city_id = c.id
		WHERE uud.id IN (
			SELECT pa.university_id FROM personalities.administrations pa WHERE pa.max_user_id = $1
			UNION
			SELECT ud.university_id FROM personalities.students ps 
			JOIN universities.university_departments ud ON ps.university_deparment_id = ud.id
			WHERE ps.max_user_id = $1
		)
	`

	rows, err := r.pool.Query(ctx, qGetAllUniversitiesForPerson, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.UniversitiesData
	for rows.Next() {
		var university models.UniversitiesData
		if err := rows.Scan(&university.ID, &university.Name, &university.City, &university.ShortName, &university.SiteUrl, &university.Description); err != nil {
			return nil, err
		}
		result = append(result, university)
	}

	return result, nil
}

func (r *PersonalitiesRepo) GetAllFacultiesForUniversity(ctx context.Context, universityID int64) ([]models.Faculties, error) {
	const qGetAllFacultiesForUniversity = `
		SELECT
			uf.id,
			uf.name,
			uud.name
		FROM universities.faculties AS uf
		JOIN universities.universities_data AS uud ON uf.university_id = uud.id
		WHERE uf.university_id = $1
		ORDER BY uf.name
	`

	rows, err := r.pool.Query(ctx, qGetAllFacultiesForUniversity, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Faculties
	for rows.Next() {
		var faculty models.Faculties
		if err := rows.Scan(&faculty.ID, &faculty.Name, &faculty.UniversityName); err != nil {
			return nil, err
		}
		result = append(result, faculty)
	}

	return result, nil
}

func (r *PersonalitiesRepo) GetAllDepartmentsForFaculty(ctx context.Context, facultyID int64) ([]models.Departments, error) {
	const qGetAllDepartmentsForFaculty = `
		SELECT
			ud_main.id,
			ud_main.name,
			COALESCE(ud_main.code, '') as code,
			uf.id,
			uf.name,
			uud.id,
			uud.name
		FROM universities.university_departments AS ud
		JOIN universities.departments AS ud_main ON ud.department_id = ud_main.id
		JOIN universities.faculties AS uf ON ud.faculty_id = uf.id
		JOIN universities.universities_data AS uud ON ud.university_id = uud.id
		WHERE ud.faculty_id = $1
		ORDER BY ud_main.name
	`

	rows, err := r.pool.Query(ctx, qGetAllDepartmentsForFaculty, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Departments
	for rows.Next() {
		var department models.Departments
		if err := rows.Scan(&department.ID, &department.Name, &department.Code, &department.FacultyID, &department.FacultyName, &department.UniversityID, &department.UniversityName); err != nil {
			return nil, err
		}
		result = append(result, department)
	}

	return result, nil
}

func (r *PersonalitiesRepo) GetAllGroupsForDepartment(ctx context.Context, departmentID int64) ([]models.Groups, error) {
	const qGetAllGroupsForDepartment = `
		SELECT 
			cg.id,
			cg.name AS group_name,
			cg.course_id,
			d.name AS department_name,
			d.code AS department_code
		FROM groups.course_groups cg
		INNER JOIN universities.courses c ON cg.course_id = c.id
		INNER JOIN universities.university_departments ud ON c.university_department_id = ud.id
		INNER JOIN universities.departments d ON ud.department_id = d.id
		WHERE 
			c.university_department_id = $1
			AND c.end_date > CURRENT_TIMESTAMP
		ORDER BY 
			cg.name
	`

	rows, err := r.pool.Query(ctx, qGetAllGroupsForDepartment, departmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Groups
	for rows.Next() {
		var group models.Groups
		if err := rows.Scan(&group.ID, &group.Name, &group.CourseID, &group.DepartmentName, &group.Code); err != nil {
			return nil, err
		}
		result = append(result, group)
	}

	return result, nil
}

func (r *PersonalitiesRepo) GetAllStudentsForGroup(ctx context.Context, groupID int64) ([]models.User, error) {
	const qGetAllStudentsForGroup = `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.is_bot,
			COALESCE(EXTRACT(EPOCH FROM u.last_activity)::int, 0) as last_activity_time,
			u.description,
			u.avatar_url,
			u.full_avatar_url
		FROM users.max_users_data AS u
		JOIN personalities.students AS s ON u.id = s.max_user_id
		WHERE s.course_group_id = $1
			AND s.is_graduated = false
		ORDER BY u.first_name, u.last_name
	`

	rows, err := r.pool.Query(ctx, qGetAllStudentsForGroup, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.UserName, &user.IsBot, &user.LastActivityTime, &user.Description, &user.AvatarUrl, &user.FullAvatarUrl); err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return result, nil
}
func (r *PersonalitiesRepo) GetAllTeachersForUniversity(ctx context.Context, universityID int64) ([]models.User, error) {
	const qGetAllTeachersForUniversity = `
		SELECT DISTINCT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.is_bot,
			COALESCE(u.last_activity, 0) as last_activity_time,
			u.description,
			u.avatar_url,
			u.full_avatar_url
		FROM users.max_users_data AS u
		JOIN personalities.teachers AS t ON u.id = t.max_user_id
		WHERE t.id IN (
			SELECT DISTINCT cgs.teacher_id 
			FROM subjects.course_group_subjects cgs
			JOIN groups.course_groups cg ON cgs.course_group_id = cg.id
			JOIN universities.courses c ON cg.course_id = c.id
			JOIN universities.university_departments ud ON c.university_department_id = ud.id
			WHERE ud.university_id = $1
			UNION
			SELECT DISTINCT egs.teacher_id
			FROM subjects.elective_group_subjects egs
			JOIN groups.elective_groups eg ON egs.elective_group_id = eg.id
			JOIN universities.semesters s ON eg.semester_id = s.id
			WHERE s.university_id = $1
		)
		ORDER BY u.first_name, u.last_name
	`

	rows, err := r.pool.Query(ctx, qGetAllTeachersForUniversity, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.UserName, &user.IsBot, &user.LastActivityTime, &user.Description, &user.AvatarUrl, &user.FullAvatarUrl); err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return result, nil
}

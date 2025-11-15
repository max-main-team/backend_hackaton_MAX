package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/schedules"
)

var ErrScheduleConflict = errors.New("schedule conflict")

type SchedulesRepo struct {
	pool *pgxpool.Pool
}

func NewScheduleRepo(pool *pgxpool.Pool) *SchedulesRepo {
	return &SchedulesRepo{pool: pool}
}
func (r *SchedulesRepo) CreateClass(ctx context.Context, class schedules.Class) (int64, error) {
	const q = `
		INSERT INTO schedules.classes (
			university_id,
			pair_number,
			start_time,
			end_time
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	var id int64
	err := r.pool.QueryRow(ctx, q,
		class.UniversityID,
		class.PairNumber,
		class.StartTime,
		class.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *SchedulesRepo) DeleteClass(ctx context.Context, class_id int64) error {
	const q = `DELETE FROM schedules.classes WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, class_id)
	return err
}

func (r *SchedulesRepo) GetClassesByUniversity(ctx context.Context, universityID int64) ([]schedules.Class, error) {
	const q = `
		SELECT id, university_id, pair_number, start_time, end_time
		FROM schedules.classes
		WHERE university_id = $1
		ORDER BY pair_number;
	`

	rows, err := r.pool.Query(ctx, q, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []schedules.Class
	for rows.Next() {
		var class schedules.Class
		if err := rows.Scan(
			&class.ID,
			&class.UniversityID,
			&class.PairNumber,
			&class.StartTime,
			&class.EndTime,
		); err != nil {
			return nil, err
		}
		result = append(result, class)
	}

	return result, nil
}

func (r *SchedulesRepo) CreateRoom(ctx context.Context, room schedules.Room) (int64, error) {
	const q = `
		INSERT INTO schedules.rooms (
			university_id,
			room
		)
		VALUES ($1, $2)
		RETURNING id;
	`

	var id int64
	err := r.pool.QueryRow(ctx, q, room.UniversityID, room.Room).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *SchedulesRepo) DeleteRoom(ctx context.Context, room_id int64) error {
	const q = `DELETE FROM schedules.rooms WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, room_id)
	return err
}

func (r *SchedulesRepo) GetRoomsByUniversity(ctx context.Context, universityID int64) ([]schedules.Room, error) {
	const q = `
		SELECT id, university_id, room
		FROM schedules.rooms
		WHERE university_id = $1
		ORDER BY room;
	`

	rows, err := r.pool.Query(ctx, q, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []schedules.Room
	for rows.Next() {
		var room schedules.Room
		if err := rows.Scan(
			&room.ID,
			&room.UniversityID,
			&room.Room,
		); err != nil {
			return nil, err
		}
		result = append(result, room)
	}

	return result, nil
}

// CreateLesson делает проверки:
// - аудитория свободна (НО лекция может пересекаться с другими лекциями);
// - преподаватель не занят (игнорируем лекции для проверки, чтобы один лекционный слот на много групп проходил);
// - группа / студенты не заняты (лекции считаются обычными занятиями);
// и потом вставляет запись в schedules.groups_schedules.
func (r *SchedulesRepo) CreateLesson(ctx context.Context, req schedules.CreateLesson) (int64, error) {
	if (req.CourseGroupSubjectID == nil && req.ElectiveGroupSubjectID == nil) ||
		(req.CourseGroupSubjectID != nil && req.ElectiveGroupSubjectID != nil) {
		return 0, errors.New("exactly one of course_group_subject_id or elective_group_subject_id must be set")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	day := string(req.Day)
	interval := string(req.Interval)

	// Получаем teacher_id, group/elective_group и тип предмета (lecture/practice/…)
	var (
		teacherID       int64
		courseGroupID   *int64
		electiveGroupID *int64
		subjectType     string
	)

	if req.CourseGroupSubjectID != nil {
		const q = `
			SELECT teacher_id, course_group_id, subject_type::text
			FROM subjects.course_group_subjects
			WHERE id = $1;
		`
		var cgID int64
		if err = tx.QueryRow(ctx, q, *req.CourseGroupSubjectID).Scan(&teacherID, &cgID, &subjectType); err != nil {
			return 0, err
		}
		courseGroupID = &cgID
	} else {
		const q = `
			SELECT teacher_id, elective_group_id, subject_type::text
			FROM subjects.elective_group_subjects
			WHERE id = $1;
		`
		var egID int64
		if err = tx.QueryRow(ctx, q, *req.ElectiveGroupSubjectID).Scan(&teacherID, &egID, &subjectType); err != nil {
			return 0, err
		}
		electiveGroupID = &egID
	}

	// 1. Проверка комнаты.
	//
	// Здесь игнорируем существующие ЛЕКЦИИ (room у лекций может совпадать),
	// но любые другие предметы участвуют в конфликте.
	const qRoomCounts = `
		SELECT
			COUNT(*) FILTER (WHERE gs."interval" = 'every two week') AS twoweek_count,
			COUNT(*) FILTER (WHERE gs."interval" <> 'every two week') AS other_count
		FROM schedules.groups_schedules gs
		LEFT JOIN subjects.course_group_subjects cgs
		       ON gs.course_group_subjet_id = cgs.id
		LEFT JOIN subjects.elective_group_subjects egs
		       ON gs.elective_group_subject_id = egs.id
		WHERE gs.day = $1
		  AND gs.class_id = $2
		  AND gs.room_id = $3
		  AND COALESCE(cgs.subject_type::text, egs.subject_type::text) <> 'lecture';
	`

	var roomTwoweek, roomOther int
	if err = tx.QueryRow(ctx, qRoomCounts, day, req.ClassID, req.RoomID).Scan(&roomTwoweek, &roomOther); err != nil {
		return 0, err
	}
	if checkIntervalConflict(interval, roomTwoweek, roomOther) {
		return 0, ErrScheduleConflict
	}

	// 2. Преподаватель.
	//
	// Тоже игнорируем лекции (одна лекция на много групп ок),
	// но преподаватель не может вести две НЕ лекции одновременно.
	const qTeacherCounts = `
		SELECT
			COUNT(*) FILTER (WHERE gs."interval" = 'every two week') AS twoweek_count,
			COUNT(*) FILTER (WHERE gs."interval" <> 'every two week') AS other_count
		FROM schedules.groups_schedules gs
		LEFT JOIN subjects.course_group_subjects cgs
		       ON gs.course_group_subjet_id = cgs.id
		LEFT JOIN subjects.elective_group_subjects egs
		       ON gs.elective_group_subject_id = egs.id
		WHERE gs.day = $1
		  AND gs.class_id = $2
		  AND (cgs.teacher_id = $3 OR egs.teacher_id = $3)
		  AND COALESCE(cgs.subject_type::text, egs.subject_type::text) <> 'lecture';
	`

	var teacherTwoWeek, teacherOther int
	if err = tx.QueryRow(ctx, qTeacherCounts, day, req.ClassID, teacherID).Scan(&teacherTwoWeek, &teacherOther); err != nil {
		return 0, err
	}
	if checkIntervalConflict(interval, teacherTwoWeek, teacherOther) {
		return 0, ErrScheduleConflict
	}

	// 3. Конфликты по группе/студентам.
	//
	// Для групп и студентов лекции считаются как обычные пары:
	// группа не может иметь две лекции одновременно.

	if courseGroupID != nil {
		// 3.1. Эта же учебная группа уже имеет обязательные пары в этот слот.
		const qGroupCounts = `
			SELECT
				COUNT(*) FILTER (WHERE gs."interval" = 'every two week') AS twoweek_count,
				COUNT(*) FILTER (WHERE gs."interval" <> 'every two week') AS other_count
			FROM schedules.groups_schedules gs
			JOIN subjects.course_group_subjects cgs
			  ON gs.course_group_subjet_id = cgs.id
			WHERE gs.day = $1
			  AND gs.class_id = $2
			  AND cgs.course_group_id = $3;
		`

		var groupTwoweek, groupOther int
		if err = tx.QueryRow(ctx, qGroupCounts, day, req.ClassID, *courseGroupID).Scan(&groupTwoweek, &groupOther); err != nil {
			return 0, err
		}
		if checkIntervalConflict(interval, groupTwoweek, groupOther) {
			return 0, ErrScheduleConflict
		}

		// 3.2. Студенты этой группы уже имеют элективы в этот слот.
		//
		// Здесь логика на уровне "есть ли уже перегруженные студенты":
		// считаем количества по слоту для элективов студентов этой группы.
		const qStudentElectiveCounts = `
			SELECT
				COUNT(*) FILTER (WHERE gs."interval" = 'every two week') AS twoweek_count,
				COUNT(*) FILTER (WHERE gs."interval" <> 'every two week') AS other_count
			FROM schedules.groups_schedules gs
			JOIN subjects.elective_group_subjects egs
			  ON gs.elective_group_subject_id = egs.id
			JOIN groups.students_elective_groups seg
			  ON egs.elective_group_id = seg.elective_group_id
			JOIN personalities.students s
			  ON seg.student_id = s.id
			WHERE gs.day = $1
			  AND gs.class_id = $2
			  AND s.course_group_id = $3;
		`

		var studElectTwoweek, studElectOther int
		if err = tx.QueryRow(ctx, qStudentElectiveCounts, day, req.ClassID, *courseGroupID).Scan(&studElectTwoweek, &studElectOther); err != nil {
			return 0, err
		}
		if checkIntervalConflict(interval, studElectTwoweek, studElectOther) {
			return 0, ErrScheduleConflict
		}
	}

	if electiveGroupID != nil {
		// 3.3. Студенты элективной группы уже имеют ОБЯЗАТЕЛЬНЫЕ пары в этот слот.
		const qStudentGroupCounts = `
			SELECT
				COUNT(*) FILTER (WHERE gs."interval" = 'every two week') AS twoweek_count,
				COUNT(*) FILTER (WHERE gs."interval" <> 'every two week') AS other_count
			FROM schedules.groups_schedules gs
			JOIN subjects.course_group_subjects cgs
			  ON gs.course_group_subjet_id = cgs.id
			JOIN personalities.students s
			  ON s.course_group_id = cgs.course_group_id
			JOIN groups.students_elective_groups seg
			  ON seg.student_id = s.id
			WHERE gs.day = $1
			  AND gs.class_id = $2
			  AND seg.elective_group_id = $3;
		`

		var studGroupTwoWeek, studGroupOther int
		if err = tx.QueryRow(ctx, qStudentGroupCounts, day, req.ClassID, *electiveGroupID).Scan(&studGroupTwoWeek, &studGroupOther); err != nil {
			return 0, err
		}
		if checkIntervalConflict(interval, studGroupTwoWeek, studGroupOther) {
			return 0, ErrScheduleConflict
		}
	}

	// 4. Вставка.
	const qInsert = `
		INSERT INTO schedules.groups_schedules (
			course_group_subjet_id,
			elective_group_subject_id,
			day,
			class_id,
			room_id,
			"interval"
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	var id int64
	err = tx.QueryRow(ctx, qInsert,
		req.CourseGroupSubjectID,
		req.ElectiveGroupSubjectID,
		day,
		req.ClassID,
		req.RoomID,
		interval,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *SchedulesRepo) DeleteLesson(ctx context.Context, lessonID int64) error {
	const q = `DELETE FROM schedules.groups_schedules WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, lessonID)
	return err
}

// GetUserSchedule — возвращает расписание по max_user_id.
// Если user — студент, вернёт пары как студента.
// Если user — преподаватель, вернёт пары как преподавателя.
// Если и то и то — всё вместе.
func (r *SchedulesRepo) GetUserSchedule(ctx context.Context, userID int64) ([]schedules.UserScheduleItem, error) {
	var result []schedules.UserScheduleItem

	// 1. Студент?
	var studentID int64
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM personalities.students WHERE max_user_id = $1`,
		userID,
	).Scan(&studentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		items, err := r.getStudentSchedule(ctx, studentID)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}

	// 2. Преподаватель?
	var teacherID int64
	err = r.pool.QueryRow(ctx,
		`SELECT id FROM personalities.teachers WHERE max_user_id = $1`,
		userID,
	).Scan(&teacherID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		items, err := r.getTeacherSchedule(ctx, teacherID)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}

	return result, nil
}

// ---- вспомогательная логика интервалов ----
//
// newInterval:
//   - "every week"
//   - "every two week"
//
// counts:
//   - twoweekCount: количество уже существующих пар с interval = 'every two week'
//   - otherCount  : количество существующих пар с interval <> 'every two week'
//
// Правила:
//   - Любой "не two week" (т.е. "every week") конфликтует, если есть хоть что-то:
//       otherCount > 0 ИЛИ twoweekCount > 0
//   - Новый "every two week":
//       конфликт, если otherCount > 0 (есть weekly)
//       конфликт, если twoweekCount >= 2 (две двухнедельные уже заняты)

func checkIntervalConflict(newInterval string, twoweekCount, otherCount int) bool {
	switch newInterval {
	case "every two week":
		if otherCount > 0 {
			return true
		}
		if twoweekCount >= 2 {
			return true
		}
	default: // "every week" (и любые другие, если останутся)
		if otherCount > 0 || twoweekCount > 0 {
			return true
		}
	}
	return false
}

func (r *SchedulesRepo) getStudentSchedule(ctx context.Context, studentID int64) ([]schedules.UserScheduleItem, error) {
	const q = `
		SELECT
			gs.id                                AS lesson_id,
			gs.day::text                         AS day,
			gs."interval"::text                  AS interval,
			c.pair_number                        AS pair_number,
			c.start_time                         AS start_time,
			c.end_time                           AS end_time,
			rms.id                               AS room_id,
			rms.room                             AS room,
			us.name                              AS subject_name,
			COALESCE(cgs.subject_type::text,
			         egs.subject_type::text)     AS subject_type,
			t.id                                 AS teacher_id,
			mud.first_name                       AS teacher_first_name,
			mud.last_name                        AS teacher_last_name
		FROM schedules.groups_schedules gs
		JOIN schedules.classes c
		  ON gs.class_id = c.id
		JOIN schedules.rooms rms
		  ON gs.room_id = rms.id
		LEFT JOIN subjects.course_group_subjects cgs
		  ON gs.course_group_subjet_id = cgs.id
		LEFT JOIN subjects.course_semester_subjects css
		  ON cgs.course_semester_subject_id = css.id
		LEFT JOIN subjects.university_subjects us
		  ON css.university_subject_id = us.id
		LEFT JOIN subjects.elective_group_subjects egs
		  ON gs.elective_group_subject_id = egs.id
		LEFT JOIN personalities.teachers t
		  ON t.id = COALESCE(cgs.teacher_id, egs.teacher_id)
		LEFT JOIN users.max_users_data mud
		  ON mud.id = t.max_user_id
		WHERE
			EXISTS (
				SELECT 1
				FROM personalities.students s
				WHERE s.id = $1
				  AND (
						(cgs.course_group_id = s.course_group_id)
						OR EXISTS (
							SELECT 1
							FROM groups.students_elective_groups seg
							WHERE seg.student_id = s.id
							  AND seg.elective_group_id = egs.elective_group_id
						)
				  )
			)
		ORDER BY gs.day, c.pair_number;
	`

	rows, err := r.pool.Query(ctx, q, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []schedules.UserScheduleItem
	for rows.Next() {
		var item schedules.UserScheduleItem
		if err := rows.Scan(
			&item.LessonID,
			&item.Day,
			&item.Interval,
			&item.PairNumber,
			&item.StartTime,
			&item.EndTime,
			&item.RoomID,
			&item.Room,
			&item.SubjectName,
			&item.SubjectType,
			&item.TeacherID,
			&item.TeacherFirstName,
			&item.TeacherLastName,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *SchedulesRepo) getTeacherSchedule(ctx context.Context, teacherID int64) ([]schedules.UserScheduleItem, error) {
	const q = `
		SELECT
			gs.id                                AS lesson_id,
			gs.day::text                         AS day,
			gs."interval"::text                  AS interval,
			c.pair_number                        AS pair_number,
			c.start_time                         AS start_time,
			c.end_time                           AS end_time,
			rms.id                               AS room_id,
			rms.room                             AS room,
			us.name                              AS subject_name,
			COALESCE(cgs.subject_type::text,
			         egs.subject_type::text)     AS subject_type,
			t.id                                 AS teacher_id,
			mud.first_name                       AS teacher_first_name,
			mud.last_name                        AS teacher_last_name
		FROM schedules.groups_schedules gs
		JOIN schedules.classes c
		  ON gs.class_id = c.id
		JOIN schedules.rooms rms
		  ON gs.room_id = rms.id
		LEFT JOIN subjects.course_group_subjects cgs
		  ON gs.course_group_subjet_id = cgs.id
		LEFT JOIN subjects.course_semester_subjects css
		  ON cgs.course_semester_subject_id = css.id
		LEFT JOIN subjects.university_subjects us
		  ON css.university_subject_id = us.id
		LEFT JOIN subjects.elective_group_subjects egs
		  ON gs.elective_group_subject_id = egs.id
		LEFT JOIN personalities.teachers t
		  ON t.id = COALESCE(cgs.teacher_id, egs.teacher_id)
		LEFT JOIN users.max_users_data mud
		  ON mud.id = t.max_user_id
		WHERE
			EXISTS (
				SELECT 1
				FROM personalities.teachers tt
				WHERE tt.id = $1
				  AND (cgs.teacher_id = tt.id OR egs.teacher_id = tt.id)
			)
		ORDER BY gs.day, c.pair_number;
	`

	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []schedules.UserScheduleItem
	for rows.Next() {
		var item schedules.UserScheduleItem
		if err := rows.Scan(
			&item.LessonID,
			&item.Day,
			&item.Interval,
			&item.PairNumber,
			&item.StartTime,
			&item.EndTime,
			&item.RoomID,
			&item.Room,
			&item.SubjectName,
			&item.SubjectType,
			&item.TeacherID,
			&item.TeacherFirstName,
			&item.TeacherLastName,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

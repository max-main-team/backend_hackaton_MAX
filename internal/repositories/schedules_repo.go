package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/schedules"
)

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

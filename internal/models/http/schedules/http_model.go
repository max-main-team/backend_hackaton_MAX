package schedules

import "time"

type CreateClassRequest struct {
	UniversityID int64     `json:"university_id"`
	PairNumber   int       `json:"pair_number"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type ClassesResponse struct {
	ID           int64     `json:"id"`
	UniversityID int64     `json:"university_id"`
	PairNumber   int       `json:"pair_number"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

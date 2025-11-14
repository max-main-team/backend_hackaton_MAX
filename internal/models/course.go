package models

import "time"

type Course struct {
	ID                   int64     `json:"id"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	UniversityDepartment int64     `json:"university_department_id"`
}

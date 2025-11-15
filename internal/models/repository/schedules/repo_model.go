package schedules

import "time"

type Class struct {
	ID           int64
	UniversityID int64
	PairNumber   int
	StartTime    time.Time
	EndTime      time.Time
}

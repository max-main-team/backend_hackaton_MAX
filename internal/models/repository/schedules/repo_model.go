package schedules

import "time"

type Class struct {
	ID           int64
	UniversityID int64
	PairNumber   int
	StartTime    time.Time
	EndTime      time.Time
}

type Room struct {
	ID           int64
	UniversityID int64
	Room         string
}

type DayType string
type IntervalType string

type CreateLesson struct {
	CourseGroupSubjectID   *int64
	ElectiveGroupSubjectID *int64
	Day                    DayType
	ClassID                int64
	RoomID                 int64
	Interval               IntervalType
}

type UserScheduleItem struct {
	LessonID         int64     `json:"lesson_id"`
	Day              string    `json:"day"`
	Interval         string    `json:"interval"`
	PairNumber       int       `json:"pair_number"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	RoomID           int64     `json:"room_id"`
	Room             string    `json:"room"`
	SubjectName      *string   `json:"subject_name,omitempty"`
	SubjectType      *string   `json:"subject_type,omitempty"`
	TeacherID        *int64    `json:"teacher_id,omitempty"`
	TeacherFirstName *string   `json:"teacher_first_name,omitempty"`
	TeacherLastName  *string   `json:"teacher_last_name,omitempty"`
}

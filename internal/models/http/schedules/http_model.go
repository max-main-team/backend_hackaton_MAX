package schedules

import (
	"time"
)

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

type CreateRoomRequest struct {
	UniversityID int64  `json:"university_id"`
	Room         string `json:"room"`
}
type RoomsResponse struct {
	ID           int64  `json:"id"`
	UniversityID int64  `json:"university_id"`
	Room         string `json:"room"`
}

type CreateLessonRequest struct {
	CourseGroupSubjectID   *int64 `json:"course_group_subject_id,omitempty"`
	ElectiveGroupSubjectID *int64 `json:"elective_group_subject_id,omitempty"`
	Day                    string `json:"day"`      // schedules.day_type
	ClassID                int64  `json:"class_id"` // schedules.classes.id
	RoomID                 int64  `json:"room_id"`  // schedules.rooms.id
	Interval               string `json:"interval"` // schedules.interval_type
}

type LessonsResponse struct {
	UserID   int64        `json:"user_id"`
	Schedule []LessonItem `json:"schedule"`
}

type LessonItem struct {
	LessonID int64  `json:"lesson_id"`
	Day      string `json:"day"`      // monday..sunday
	Interval string `json:"interval"` // every week / every two week

	PairNumber int       `json:"pair_number"`
	StartTime  time.Time `json:"start_time"` // HH:MM
	EndTime    time.Time `json:"end_time"`   // HH:MM

	RoomID int64  `json:"room_id"`
	Room   string `json:"room"`

	SubjectName *string `json:"subject_name,omitempty"`
	SubjectType *string `json:"subject_type,omitempty"` // lecture/practice/etc

	TeacherID        *int64  `json:"teacher_id,omitempty"`
	TeacherFirstName *string `json:"teacher_first_name,omitempty"`
	TeacherLastName  *string `json:"teacher_last_name,omitempty"`
}

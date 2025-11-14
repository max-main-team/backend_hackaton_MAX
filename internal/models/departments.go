package models

type Departments struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	FacultyID      int64  `json:"faculty_id"`
	FacultyName    string `json:"faculty_name"`
	UniversityID   int64  `json:"university_id"`
	UniversityName string `json:"university_name"`
}

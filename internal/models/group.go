package models

type Group struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Groups struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	CourseID       int64  `json:"course_id"`
	Code           string `json:"code"`
	DepartmentName string `json:"department_name"`
}

package dto

type GroupInfoResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	CourseID       int64  `json:"course_id"`
	Code           string `json:"code"`
	DepartmentName string `json:"department_name"`
}

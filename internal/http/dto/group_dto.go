package dto

type GroupInfoResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	CourseID       int64  `json:"course_id"`
	Code           string `json:"code"`
	DepartmentName string `json:"department_name"`
}

type CreateGroupRequest struct {
	GroupName    string `json:"group_name" validate:"required"`
	DepartmentID int64  `json:"department_id" validate:"required"`
	FacultyID    int64  `json:"faculty_id" validate:"required"`
	UniversityID int64  `json:"university_id" validate:"required"`
}

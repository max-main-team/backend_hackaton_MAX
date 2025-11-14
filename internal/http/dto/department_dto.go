package dto

type DepartmentInfoResponse struct {
	ID          int    `json:"id" validate:"required" example:"123456789"`
	Name        string `json:"department_name" validate:"required" example:"Sowtware Engineering"`
	Code        string `json:"code" validate:"required" example:"09.03.02"`
	FacultyName string `json:"faculty_name" validate:"required" example:"FITIP"`
}

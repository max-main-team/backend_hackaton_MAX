package dto

type CreateNewFacultyRequest struct {
	Name string `json:"faculty_name" validate:"required" example:"FITIP"`
}

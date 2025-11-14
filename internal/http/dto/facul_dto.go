package dto

type CreateNewFacultyRequest struct {
	Name string `json:"faculty_name" validate:"required" example:"FITIP"`
}

type FacultyInfoResponse struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	UniversityName string `json:"university_name"`
}

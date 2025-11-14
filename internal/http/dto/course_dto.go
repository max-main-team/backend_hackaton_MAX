package dto

type CreateCourseRequest struct {
	StartDate            string `json:"start_date" validate:"required"`
	EndDate              string `json:"end_date" validate:"required"`
	UniversityDepartment int64  `json:"university_department_id" validate:"required"`
}

type CourseInfoResponse struct {
	ID                   int64  `json:"id"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	UniversityDepartment int64  `json:"university_department_id"`
}

package personalities

import "github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"

type RequestAccessToUniversity struct {
	UniversityID int64                  `json:"university_id"`
	UserType     personalities.RoleType `json:"role"`
}

type AccessRequestResponse struct {
	Data []struct {
		RequestID int64                  `json:"request_id"`
		UserID    int64                  `json:"user_id"`
		UserType  personalities.RoleType `json:"role"`
		FirstName string                 `json:"first_name"`
		LastName  *string                `json:"last_name,omitempty"`
		Username  *string                `json:"username,omitempty"`
	} `json:"data"`
	HasMore bool `json:"has_more"`
}

type AcceptAccessRequest struct {
	UserID                 int64                  `json:"user_id" validate:"required"`
	UserType               personalities.RoleType `json:"role" validate:"required"`
	UniversityID           *int64                 `json:"university_id,omitempty"`
	FacultyID              *int64                 `json:"faculty_id,omitempty"`
	UniversityDepartmentID *int64                 `json:"university_department_id,omitempty"`
	CourseGroupID          *int64                 `json:"course_group_id,omitempty"`
}

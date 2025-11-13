package personalities

import "github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"

type RequestAccessToUniversity struct {
	UniversityID int64                  `json:"university_id"`
	UserType     personalities.RoleType `json:"role"`
}

type AccessRequestResponse struct {
	Data []struct {
		UserID   int64                  `json:"user_id"`
		UserType personalities.RoleType `json:"role"`
	} `json:"data"`
	HasMore bool `json:"has_more"`
}

package dto

type LoginResponse struct {
	AccessToken string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User        User     `json:"user" validate:"required"`
	UserRoles   []string `json:"user_roles"`
}

type WebAppInitData struct {
	QueryID    string `json:"query_id" validate:"required" example:"unique_session_id"`
	AuthDate   int    `json:"auth_date" validate:"required" example:"1633038072"`
	Hash       string `json:"hash" validate:"required" example:"abc123def456"`
	StartParam string `json:"start_param" example:"start_parameter"`

	User `json:"user" validate:"required"`

	Chat struct {
		ID   int    `json:"id" validate:"required" example:"-1001234567890"`
		Type string `json:"type" validate:"required" example:"group"`
	} `json:"chat,omitempty"`
}

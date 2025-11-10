package dto

type LoginRequest struct {
	InitData string `json:"username" validate:"required,min=3" example:"admin"`
	Password string `json:"password" validate:"required,min=8" example:"strongpassword"`
}
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type WebAppInitData struct {
	QueryID    string `json:"query_id" validate:"required" example:"unique_session_id"`
	AuthDate   int    `json:"auth_date" validate:"required" example:"1633038072"`
	Hash       string `json:"hash" validate:"required" example:"abc123def456"`
	StartParam string `json:"start_param" example:"start_parameter"`

	User struct {
		ID           int    `json:"id" validate:"required" example:"123456789"`
		FirstName    string `json:"first_name" validate:"required" example:"John"`
		LastName     string `json:"last_name" example:"Doe"`
		Username     string `json:"username" example:"johndoe"`
		LanguageCode string `json:"language_code" example:"ru"`
		PhotoURL     string `json:"photo_url" example:"https://example.com/photo.jpg"`
	} `json:"user" validate:"required"`

	Chat struct {
		ID   int    `json:"id" validate:"required" example:"-1001234567890"`
		Type string `json:"type" validate:"required" example:"group"`
	} `json:"chat,omitempty"`
}

package dto

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type User struct {
	ID           int    `json:"id" validate:"required" example:"123456789"`
	FirstName    string `json:"first_name" validate:"required" example:"John"`
	LastName     string `json:"last_name" example:"Doe"`
	Username     string `json:"username" example:"johndoe"`
	LanguageCode string `json:"language_code" example:"ru"`
	PhotoURL     string `json:"photo_url" example:"https://example.com/photo.jpg"`
}

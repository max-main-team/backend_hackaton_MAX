package dto

type CreateEventRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	PhotoUrl    string `json:"photo_url" validate:"required"`
}

type EventResponse struct {
	ID           int64  `json:"id"`
	UniversityID int64  `json:"university_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PhotoUrl     string `json:"photo_url"`
}

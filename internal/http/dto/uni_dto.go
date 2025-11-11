package dto

type UniInfoResponse struct {
	ID        int    `json:"id" validate:"required" example:"123456789"`
	Name      string `json:"uni_name" validate:"required" example:"ITMO University"`
	ShortName string `json:"uni_short_name" example:"ITMO"`
	City      string `json:"city" validate:"required" example:"Saint-Petersburg"`
}

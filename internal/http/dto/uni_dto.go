package dto

type UniInfoResponse struct {
	ID          int    `json:"id" validate:"required" example:"123456789"`
	Name        string `json:"uni_name" validate:"required" example:"ITMO University"`
	City        string `json:"city" validate:"required" example:"Saint-Petersburg"`
	ShortName   string `json:"uni_short_name" example:"ITMO"`
	SiteUrl     string `json:"site_url" example:"https://itmo.ru"`
	Description string `json:"description" example:"One of the leading Russian universities in the field of information technology, optical design, and engineering."`
}

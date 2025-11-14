package dto

type UniInfoResponse struct {
	ID          int    `json:"id" validate:"required" example:"123456789"`
	Name        string `json:"uni_name" validate:"required" example:"ITMO University"`
	City        string `json:"city" validate:"required" example:"Saint-Petersburg"`
	ShortName   string `json:"uni_short_name" example:"ITMO"`
	SiteUrl     string `json:"site_url" example:"https://itmo.ru"`
	Description string `json:"description" example:"One of the leading Russian universities in the field of information technology, optical design, and engineering."`
}

type SemesterPeriod struct {
	StartDate string `json:"start_date" validate:"required" example:"2005-12-23T00:00:00Z"`
	EndDate   string `json:"end_date" validate:"required" example:"2006-12-23T00:00:00Z"`
}

type CreateSemestersRequest struct {
	ID      int              `json:"uni_id" validate:"required"`
	Periods []SemesterPeriod `json:"periods" validate:"required,dive"`
}

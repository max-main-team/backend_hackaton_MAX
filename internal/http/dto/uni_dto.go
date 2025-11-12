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
	StartDate int64 `json:"start_date" validate:"required"`
	EndDate   int64 `json:"end_date" validate:"required"`
}

type CreateSemestersRequest struct {
	ID      int              `json:"uni_id" validate:"required"`
	Periods []SemesterPeriod `json:"periods" validate:"required,dive"`
}

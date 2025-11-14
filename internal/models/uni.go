package models

import "time"

type UniversitiesData struct {
	ID          int
	Name        string
	City        string
	ShortName   string
	SiteUrl     *string
	Description *string
	PhotoUrl    *string
}
type SemesterPeriod struct {
	StartDate time.Time
	EndDate   time.Time
}

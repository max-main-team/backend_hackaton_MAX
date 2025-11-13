package subjects

type CreateSubjectRequest struct {
	Name         string `json:"name"`
	UniversityID int64  `json:"university_id"`
}

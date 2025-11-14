package subjects

type CreateSubjectRequest struct {
	Name         string `json:"name"`
	UniversityID int64  `json:"university_id"`
}

type GetSubjectsRequest struct {
	UniversityID int64 `json:"university_id"`
}

type SubjectsResponse struct {
	Data []struct {
		Name string `json:"name"`
		ID   int64  `json:"id"`
	} `json:"data"`
	HasMore bool `json:"has_more"`
}

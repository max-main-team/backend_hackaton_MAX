package personalities

type RoleType string

const (
	Student RoleType = "student"
	Teacher RoleType = "teacher"
	Admin   RoleType = "administration"
)

type RequestAccessToUniversity struct {
	UniversityID int64    `json:"university_id"`
	UserType     RoleType `json:"role"`
}

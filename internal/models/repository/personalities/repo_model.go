package personalities

type RoleType string

const (
	Student RoleType = "student"
	Teacher RoleType = "teacher"
	Admin   RoleType = "administration"
)

type UniversityAccess struct {
	UserID       int64
	UserType     RoleType
	UniversityID int64
}

type PaginationParams struct {
	Limit  int64
	Offset int64
}

type AccessRequests struct {
	Requests []struct {
		UserID   int64
		UserType RoleType
	}
}

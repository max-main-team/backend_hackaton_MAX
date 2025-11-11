package models

type User struct {
	ID               int64
	FirstName        string
	LastName         *string
	UserName         *string
	IsBot            bool
	LastActivityTime int
	Description      *string
	AvatarUrl        *string
	FullAvatarUrl    *string
}

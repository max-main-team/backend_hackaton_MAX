package models

type User struct {
	ID               int
	FirstName        string
	LastName         *string
	UserName         *string
	IsBot            bool
	LastActivityTime int
	Description      *string
	AvatarUrl        *string
	FullAvatarUrl    *string
}

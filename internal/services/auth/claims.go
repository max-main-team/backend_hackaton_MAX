package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	ID             int    `json:"ID"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	UserName       string `json:"username"`
	IsBot          bool   `json:"is_bot"`
	LastAstiveName int    `json:"last_active_name"`
	Description    string `json:"description"`
	AvatarUrl      string `json:"avatar_url"`
	FullAvatarUrl  string `json:"full_avatar_url"`
	jwt.RegisteredClaims
}

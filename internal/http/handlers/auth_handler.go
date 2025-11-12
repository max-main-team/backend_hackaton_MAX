package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services/auth"
	"github.com/vmkteam/embedlog"
)

var secure bool = true

type AuthHandler struct {
	jwtService  *auth.JWTService
	userRepo    repositories.UserRepository
	refreshRepo repositories.RefreshTokenRepository
	botToken    string
}

func NewAuthHandler(
	jwt *auth.JWTService,
	uRepo repositories.UserRepository,
	rRepo repositories.RefreshTokenRepository,
	bToken string,
) *AuthHandler {
	return &AuthHandler{
		jwtService:  jwt,
		userRepo:    uRepo,
		refreshRepo: rRepo,
		botToken:    bToken,
	}
}

type ErrorResponse struct {
	Message string `json:"message" example:"error description"`
	Error   string `json:"error"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (h *AuthHandler) Login(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	if err := c.Request().ParseForm(); err != nil {
		log.Errorf("Failed to parse form: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	authDate := c.Request().FormValue("auth_date")

	hash := c.Request().FormValue("hash")
	userStr := c.Request().FormValue("user")

	if authDate == "" || hash == "" {
		log.Errorf("Missing required fields: auth_date=%s, hash=%s", authDate, hash)
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields")
	}

	var userData struct {
		ID        int     `json:"id"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Username  *string `json:"username"`
		PhotoURL  *string `json:"photo_url"`
	}

	if userStr != "" {
		if err := json.Unmarshal([]byte(userStr), &userData); err != nil {
			log.Errorf("Failed to unmarshal user data: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user data format")
		}
	}

	if !h.validateMAXData(c.Request().Form, hash) {
		log.Errorf("Invalid MAX WebApp data")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid init data")
	}

	log.Printf("Finding user in DB with ID: %d", userData.ID)

	var user *models.User

	user, err := h.userRepo.GetUserByID(context.TODO(), int64(userData.ID))
	if err != nil {
		log.Errorf("Failed find user in db. err: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {

			newUser := &models.User{
				ID:            int64(userData.ID),
				FirstName:     userData.FirstName,
				LastName:      &userData.LastName,
				UserName:      userData.Username,
				IsBot:         false,
				AvatarUrl:     userData.PhotoURL,
				FullAvatarUrl: userData.PhotoURL,
			}

			err := h.userRepo.CreateNewUser(context.TODO(), newUser)
			if err != nil {
				log.Errorf("Failed to create new user. err: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new user")
			}
			user = newUser
			log.Printf("Created new user with ID: %d", user.ID)
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to find user. err: "+err.Error())
	}

	access, refresh, err := h.jwtService.GenerateTokenPair(
		int(user.ID),
		user.LastActivityTime,
		user.FirstName,
		user.LastName,
		user.UserName,
		user.Description,
		user.AvatarUrl,
		user.FullAvatarUrl,
		user.IsBot,
	)
	if err != nil {
		log.Errorf("Token generation error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	expires := time.Now().Add(h.jwtService.RefreshExpiry())

	rt := &models.RefreshToken{
		UserID:    int(user.ID),
		Token:     refresh,
		ExpiresAt: expires,
		CreatedAt: time.Now(),
	}

	if err := h.refreshRepo.Save(rt); err != nil {
		log.Printf("Refresh token save error: %v", err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})

	userRoles, err := h.userRepo.GetUserRolesByID(context.TODO(), user.ID)
	if err != nil {
		log.Errorf("Failed find user roles: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed find user roles")
	}

	responseUser := dto.User{
		ID:        userData.ID,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Username:  NewString(userData.Username),
		PhotoURL:  NewString(userData.PhotoURL),
	}

	log.Printf("User %d logged in successfully", userData.ID)
	log.Printf("AccessToken: %s", access)

	response := dto.LoginResponse{
		AccessToken: access,
		User:        responseUser,
		UserRoles:   userRoles.Roles,
	}

	// Логируем что будем отправлять
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("❌ JSON marshal error: %v", err)
	} else {
		log.Printf("📤 Sending JSON: %s", string(jsonBytes))
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: access,
		User:        responseUser,
		UserRoles:   userRoles.Roles,
	})
}

func NewString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func (h *AuthHandler) validateMAXData(form url.Values, receivedHash string) bool {

	var dataCheckStrings []string

	for key, values := range form {

		if key == "hash" {
			continue
		}

		if len(values) > 0 {
			dataCheckStrings = append(dataCheckStrings, fmt.Sprintf("%s=%s", key, values[0]))
		}
	}

	sort.Strings(dataCheckStrings)

	dataCheckString := strings.Join(dataCheckStrings, "\n")

	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(h.botToken))
	secretKey := mac.Sum(nil)

	mac = hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(mac.Sum(nil))

	return calculatedHash == receivedHash
}

func (h *AuthHandler) Refresh(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		log.Errorf("Refresh cookie error. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Refresh cookie error")
	}
	refreshRaw := cookie.Value
	rt, err := h.refreshRepo.Find(refreshRaw)
	if err != nil {
		log.Errorf("Invalid refresh token: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid refresh token")
	}

	if rt.ExpiresAt.Before(time.Now()) {
		log.Errorf("Expired refresh token. err: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Expired refresh token")
	}

	if err := h.refreshRepo.Delete(refreshRaw); err != nil {
		log.Errorf("Refresh token delete error: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Refresh token delete error")
	}

	user, err := h.userRepo.GetUserByID(context.TODO(), int64(rt.UserID))
	if err != nil {
		log.Errorf("User lookup error: %d", rt.UserID)
		return echo.NewHTTPError(http.StatusInternalServerError, "System error")
	}

	access, refresh, err := h.jwtService.GenerateTokenPair(
		int(user.ID),
		user.LastActivityTime,
		user.FirstName,
		user.LastName,
		user.UserName,
		user.Description,
		user.AvatarUrl,
		user.FullAvatarUrl,
		user.IsBot,
	)
	if err != nil {
		log.Errorf("Token generation error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	expires := time.Now().Add(h.jwtService.RefreshExpiry())
	uid := user.ID

	newRT := &models.RefreshToken{
		UserID:    int(uid),
		Token:     refresh,
		ExpiresAt: expires,
	}

	if err := h.refreshRepo.Save(newRT); err != nil {
		log.Printf("Refresh token save error: %v", err)
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})
	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: access,
	})
}

type TokenCheckResponse struct {
	AccessToken  any
	RefreshToken any
	User         any
}

type TokenStatus struct {
	Valid bool
}

type UserInfo struct {
	Username  string
	FirstName string
	UserID    int
}

func (h *AuthHandler) CheckToken(c echo.Context) error {
	user := auth.GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "Invalid or expired access token",
		})
	}
	refreshValid := false
	refreshCookie, err := c.Cookie("refresh_token")
	if err == nil && refreshCookie.Value != "" {
		storedToken, err := h.refreshRepo.Find(refreshCookie.Value)
		if err == nil {
			refreshValid = storedToken.ExpiresAt.After(time.Now())
			uid := user.ID
			if storedToken.UserID != int(uid) {
				refreshValid = false
			}
		}
	}

	username := ""
	if user.UserName != nil {
		username = *user.UserName
	}

	response := TokenCheckResponse{
		AccessToken: TokenStatus{
			Valid: true,
		},
		RefreshToken: TokenStatus{
			Valid: refreshValid,
		},
		User: UserInfo{
			Username:  username,
			FirstName: user.FirstName,
			UserID:    int(user.ID),
		},
	}

	return c.JSON(http.StatusOK, response)
}

package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/max-main-team/backend_hackaton_MAX/internal/repositories"
	"github.com/max-main-team/backend_hackaton_MAX/internal/services/auth"
	"github.com/vmkteam/embedlog"
)

var secure bool = false

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

	// Логируем ВЕСЬ запрос
	log.Printf("=== FULL REQUEST ===")
	log.Printf("Method: %s", c.Request().Method)
	log.Printf("URL: %s", c.Request().URL.String())
	log.Printf("Headers: %v", c.Request().Header)

	// Тело запроса
	body, _ := io.ReadAll(c.Request().Body)
	log.Printf("Body: %s", string(body))

	// Восстанавливаем body для дальнейшей обработки
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	log.Printf("=== END REQUEST ===")

	var req dto.WebAppInitData
	if err := c.Bind(&req); err != nil {
		log.Errorf("Invalid format for WebAppInitData. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid request format")
	}

	if !auth.ValidateInitData(&req, h.botToken) {
		log.Errorf("Invalid WebAppInitData: %+v", req)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid init data")
	}

	user, err := h.userRepo.GetUserByID(context.TODO(), req.ID)
	if err != nil {
		log.Errorf("Failed find user in db. err: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to find user. err: "+err.Error())
	}

	access, refresh, err := h.jwtService.GenerateTokenPair(user.ID, user.LastActivityTime, user.FirstName, user.LastName, user.UserName, user.Description, user.AvatarUrl, user.FullAvatarUrl, user.IsBot)
	if err != nil {
		log.Errorf("Token generation error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}
	expires := time.Now().Add(h.jwtService.RefreshExpiry())

	uid := user.ID

	rt := &models.RefreshToken{
		UserID:    uid,
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
		log.Errorf("ailed find user roled %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed find user roled")
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: access,
		User:        req.User,
		UserRoles:   userRoles.Roles,
	})
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

	user, err := h.userRepo.GetUserByID(context.TODO(), rt.UserID)
	if err != nil {
		log.Errorf("User lookup error: %d", rt.UserID)
		return echo.NewHTTPError(http.StatusInternalServerError, "System error")
	}

	access, refresh, err := h.jwtService.GenerateTokenPair(
		user.ID,
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
		UserID:    uid,
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
			if storedToken.UserID != uid {
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
			UserID:    user.ID,
		},
	}

	return c.JSON(http.StatusOK, response)
}

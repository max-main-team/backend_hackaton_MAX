package handlers

import (
	"context"
	"errors"
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

// ErrorResponse стандартный формат ошибки
type ErrorResponse struct {
	Message string `json:"message" example:"error description"`
	Error   string `json:"error"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (h *AuthHandler) Login(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	var req dto.WebAppInitData
	if err := c.Bind(&req); err != nil {
		log.Errorf("Invalid format for WebAppInitData. err: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if !auth.ValidateInitData(&req, h.botToken) {
		log.Errorf("Invalid WebAppInitData: %+v", req)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid init data")
	}

	user, err := h.userRepo.GetUserByID(context.TODO(), req.User.ID)
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
	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: access,
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

// // CreateUser godoc
// // @Summary Создание нового пользователя (только для администраторов)
// // @Description Создаёт нового пользователя. Требуется авторизация и роль admin.
// // @Tags users
// // @Security ApiKeyAuth
// // @Accept json
// // @Produce json
// // @Param user body RegisterRequest true "Данные пользователя"
// // @Success 201 {object} map[string]string "Возвращает информацию о созданном пользователе (message, id)"
// // @Failure 400 {object} ErrorResponse "Неверный формат запроса или неверная роль"
// // @Failure 403 {object} ErrorResponse "Требуется права администратора"
// // @Failure 409 {object} ErrorResponse "Имя пользователя уже занято"
// // @Failure 500 {object} ErrorResponse "Ошибка сервера при создании пользователя"
// // @Router /admin/users [post]
// func (h *AuthHandler) CreateUser(c echo.Context) error {
// 	// Проверка прав доступа

// 	log.Println("CreateUser called")
// 	log.Println("Context user:", c.Get("user"))
// 	log.Println(c.Get("user").(*model.User).Role)
// 	currentUser, ok := c.Get("user").(*model.User)
// 	if !ok || currentUser.Role != "admin" {
// 		return echo.NewHTTPError(http.StatusForbidden, "Admin access required")
// 	}

// 	var req RegisterRequest
// 	if err := c.Bind(&req); err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
// 	}

// 	if _, err := h.userRepo.FindByUsername(req.Username); err == nil {
// 		return echo.NewHTTPError(http.StatusConflict, "Username already exists")
// 	}

// 	//VehicleType := get
// 	val, err := strconv.Atoi(req.ThresholdValue)
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid threshold value")
// 	}
// 	user := &model.User{
// 		Username:       req.Username,
// 		Password:       req.Password,
// 		Role:           req.Role,
// 		VehicleType:    req.VehicleType,
// 		ThresholdValue: val,
// 	}

// 	log.Println("user to create:", user)
// 	if err := user.HashPassword(); err != nil {
// 		log.Printf("Password hashing error: %v", err)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "User creation failed")
// 	}

// 	if err := h.userRepo.Create(user); err != nil {
// 		log.Printf("User creation error: %v", err)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "User creation failed")
// 	}

// 	return c.JSON(http.StatusCreated, map[string]string{
// 		"message": "User created successfully",
// 		"id":      user.ID,
// 	})
// }

// // Logout godoc
// // @Summary Выход из системы
// // @Description Инвалидирует refresh token (по значению из cookie) и удаляет cookie на клиенте.
// // @Tags auth
// // @Accept json
// // @Produce json
// // @Success 200 {object} map[string]string "Возвращает пустой объект при успешном выходе"
// // @Failure 400 {object} ErrorResponse "Не удалось прочитать cookie или токен неизвестен"
// // @Failure 500 {object} ErrorResponse "Ошибка сервера при удалении токена"
// // @Router /user/logout [post]
// func (h *AuthHandler) Logout(c echo.Context) error {

// 	cookie, err := c.Cookie("refresh_token")
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Logout cookie error")
// 	}
// 	value := cookie.Value
// 	if _, err := h.refreshRepo.Find(value); err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read cookie")
// 	}

// 	if err := h.refreshRepo.Delete(value); err != nil {
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete cookie")
// 	}
// 	cookie = &http.Cookie{
// 		Name:     "refresh_token",
// 		Value:    "",
// 		Path:     "/",
// 		HttpOnly: true,
// 		Secure:   secure,
// 		SameSite: http.SameSiteLaxMode,
// 		Expires:  time.Unix(0, 0),
// 		MaxAge:   -1,
// 	}
// 	http.SetCookie(c.Response(), cookie)

// 	return c.JSON(http.StatusOK, map[string]string{})
// }

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

	response := TokenCheckResponse{
		AccessToken: TokenStatus{
			Valid: true,
		},
		RefreshToken: TokenStatus{
			Valid: refreshValid,
		},
		User: UserInfo{
			Username:  *user.UserName,
			FirstName: user.FirstName,
			UserID:    user.ID,
		},
	}

	return c.JSON(http.StatusOK, response)
}

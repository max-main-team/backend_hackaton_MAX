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

// Login godoc
// @Summary      User login via MAX WebApp
// @Description  Authenticate user using MAX WebApp init data and return JWT tokens
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        auth_date  formData  string  true  "Authentication date"
// @Param        hash       formData  string  true  "Authentication hash"
// @Param        user       formData  string  false "User data JSON"
// @Success      200        {object}  dto.LoginResponse  "JWT tokens"
// @Failure      400        {object}  echo.HTTPError     "Invalid request data"
// @Failure      401        {object}  echo.HTTPError     "Invalid init data"
// @Failure      500        {object}  echo.HTTPError     "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	if err := c.Request().ParseForm(); err != nil {
		log.Errorf("[Login] Failed to parse form: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	authDate := c.Request().FormValue("auth_date")

	hash := c.Request().FormValue("hash")
	userStr := c.Request().FormValue("user")

	if authDate == "" || hash == "" {
		log.Errorf("[Login] Missing required fields: auth_date=%s, hash=%s", authDate, hash)
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
			log.Errorf("[Login] Failed to unmarshal user data: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user data format")
		}
	}

	if !h.validateMAXData(c.Request().Form, hash) {
		log.Errorf("[Login] Invalid MAX WebApp data")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid init data")
	}

	var user *models.User

	user, err := h.userRepo.GetUserByID(context.TODO(), int64(userData.ID))
	if err != nil {
		log.Errorf("[Login] Failed find user in db. err: %v", err)
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
				log.Errorf("[Login] Failed to create new user. err: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new user")
			}
			user = newUser
			log.Printf("[Login] Created new user with ID: %d", user.ID)
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "Failed to find user. err: "+err.Error())
		}
	} else {
		// Пользователь существует - обновляем его данные
		user.FirstName = userData.FirstName
		user.LastName = &userData.LastName
		user.UserName = userData.Username
		user.AvatarUrl = userData.PhotoURL
		user.FullAvatarUrl = userData.PhotoURL

		err = h.userRepo.UpdateUser(context.TODO(), user)
		if err != nil {
			log.Errorf("[Login] Failed to update user data. err: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user data")
		}
		log.Printf("[Login] Updated user data for ID: %d", user.ID)
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
		log.Errorf("[Login] Token generation error: %v", err)
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
		log.Printf("[Login] Refresh token save error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		Expires:  expires,
	})

	userRoles, err := h.userRepo.GetUserRolesByID(context.TODO(), user.ID)
	if err != nil {
		log.Errorf("[Login] Failed find user roles: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed find user roles")
	}

	responseUser := dto.User{
		ID:        userData.ID,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Username:  NewString(userData.Username),
		PhotoURL:  NewString(userData.PhotoURL),
	}

	log.Printf("[Login] User id: %d, name: %v logged in successfully. AccessToken: %v  ", userData.ID, userData.FirstName, access)

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

// Refresh godoc
// @Summary      Refresh JWT tokens
// @Description  Refresh access and refresh tokens using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RefreshRequest      true  "Refresh token"
// @Success      200      {object}  dto.LoginResponse   "New JWT tokens"
// @Failure      400      {object}  echo.HTTPError      "Invalid request body"
// @Failure      401      {object}  echo.HTTPError      "Invalid or expired refresh token"
// @Failure      500      {object}  echo.HTTPError      "Internal server error"
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {

	log := c.Get("logger").(embedlog.Logger)

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		log.Errorf("[Refresh] Refresh cookie error. err: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Refresh cookie error")
	}
	refreshRaw := cookie.Value
	rt, err := h.refreshRepo.Find(refreshRaw)
	if err != nil {
		log.Errorf("[Refresh] Invalid refresh token: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid refresh token")
	}

	if rt.ExpiresAt.Before(time.Now()) {
		log.Errorf("[Refresh] Expired refresh token. err: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Expired refresh token")
	}

	if err := h.refreshRepo.Delete(refreshRaw); err != nil {
		log.Errorf("[Refresh] Refresh token delete error: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Refresh token delete error")
	}

	user, err := h.userRepo.GetUserByID(context.TODO(), int64(rt.UserID))
	if err != nil {
		log.Errorf("[Refresh] User lookup error: %d", rt.UserID)
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
		log.Errorf("[Refresh] Token generation error: %v", err)
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
		log.Printf("[Refresh] Refresh token save error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
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

// CheckToken godoc
// @Summary      Check JWT token validity
// @Description  Verify if the provided JWT access token is valid
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "status: token is valid"
// @Failure      401  {object}  echo.HTTPError     "Unauthorized - invalid or missing token"
// @Failure      500  {object}  echo.HTTPError     "Internal server error"
// @Router       /auth/checkToken [get]
// @Security     BearerAuth
func (h *AuthHandler) CheckToken(c echo.Context) error {
	log := c.Get("logger").(embedlog.Logger)

	log.Print(context.Background(), "[CheckToken] CheckToken called")
	user := auth.GetUserFromContext(c)
	if user == nil {
		log.Printf("User not found in context")
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

// Авторизация:  ✅(Миша)

// Администрация:
// 1) Просмотор заявок на вступление GET         					✅(Артем)
// 2) Рассмотрение заявок на вступление POST     					⛔️(Артем)
// 4) Создание новых семестров POST			  					✅(Миша)
// 5) Создание новых предметов POST			 					✅(Миша)
// 6) Создание новых групп POST				  					⛔️(Артем)
// 7) Назначение преподавателей на предметы и группы POST 			⛔️(?)
// 8) Добавление мероприятий POST 									⛔️(?)
// 9) Составление расписания POST 									⛔️(?)

// Абитуриенты:
// 1) Просмтор ВУЗОВ GET    										 ✅(Миша)
// 2) Запрос на получение доступа(админ, учитель, семестров) POST.  ✅(Артем)

// Студенты:
// 1) Просмтор оценок GET  										⛔️(?)
// 2) Просмтор расписания GET   									⛔️(?)
// 3) Просмтор мероприятий GET   									⛔️(?)
// 4) Просмтор информации о ВУЗе GET   							⛔️(?)

// Преподаватели:
// 1) Просмтор информации о ВУЗе GET   							⛔️(?)
// 2) Просмтор расписания GET   									⛔️(?)
// 3) Добавление оценок POST  									    ⛔️(?)

// Общее для преподавателей и студентов:
// Персоналити:
// 0) Выбор вуза (для случаев, если у студентов 2 вуза)  		POST personalities/uni				☑️(Миша)
// 1) Просмотр факультетов для конкретного вуpf ET              ☑️(Миша)
// 2) Просмотр направлений GET            ☑️(Миша)
// 3) Просмотр групп GET                         ☑️(Миша)
// 4) Просмотр студентов групп GET      ☑️(Миша)

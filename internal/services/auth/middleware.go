package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
)

const UserKey = "user"

func (s *JWTService) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {

				reqID = fmt.Sprintf("%d", time.Now().UnixNano())
				c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			}

			log.Printf("[req %s] -> %s %s from=%s", reqID, c.Request().Method, c.Request().RequestURI, c.RealIP())

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[req %s] auth missing", reqID)
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("[req %s] invalid auth format", reqID)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			tokenString := parts[1]
			claims, err := s.ParseToken(tokenString)
			if err != nil {
				log.Printf("[req %s] token parse failed: %v", reqID, err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: "+err.Error())
			}
			user := &models.User{
				ID:               int64(claims.ID),
				FirstName:        claims.FirstName,
				LastName:         &claims.LastName,
				UserName:         &claims.UserName,
				IsBot:            claims.IsBot,
				LastActivityTime: claims.LastAstiveName,
				Description:      &claims.Description,
				AvatarUrl:        &claims.AvatarUrl,
				FullAvatarUrl:    &claims.FullAvatarUrl,
			}
			c.Set(UserKey, user)

			log.Printf("authenticated user_id = %v", user.ID)

			return next(c)
		}
	}
}

func GetUserFromContext(c echo.Context) *models.User {
	if user, ok := c.Get(UserKey).(*models.User); ok {
		return user
	}
	return nil
}

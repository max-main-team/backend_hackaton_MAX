package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	"github.com/vmkteam/embedlog"
)

const UserKey = "user"

func (s *JWTService) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			log := c.Get("logger").(embedlog.Logger)

			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = fmt.Sprintf("%d", time.Now().UnixNano())
				c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			}

			method := c.Request().Method
			path := c.Request().RequestURI

			log.Printf("[JWTMiddleware] AUTH_START %s %s", method, path)

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[JWTMiddleware] AUTH_FAIL %s %s reason=missing_auth_header", method, path)
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Errorf("[JWTMiddleware] AUTH_FAIL %s %s reason=invalid_auth_format", method, path)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			tokenString := parts[1]
			claims, err := s.ParseToken(tokenString)
			if err != nil {
				log.Errorf("[JWTMiddleware] AUTH_FAIL token parse failed: %v. token input: %v", err, tokenString)
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

			log.Printf("[JWTMiddleware] AUTH_SUCCESS %s %s max_id_user=%d", method, path, claims.ID)
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

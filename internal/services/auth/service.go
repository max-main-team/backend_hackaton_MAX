package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	config "github.com/max-main-team/backend_hackaton_MAX/cfg"
	"github.com/max-main-team/backend_hackaton_MAX/internal/http/dto"
)

type JWTService struct {
	secret []byte
	Expiry time.Duration
}

func (s *JWTService) RefreshExpiry() time.Duration {
	return s.Expiry * 24
}
func NewJWTService(cfg config.Config) *JWTService {

	if cfg.AuthConfig.JWTSecret == "" {
		log.Fatal("JWT secret is empty! Check your config file")
	}
	log.Printf("JWTService created with secret length: %d", len(cfg.AuthConfig.JWTSecret))

	return &JWTService{
		secret: []byte(cfg.AuthConfig.JWTSecret),
		Expiry: time.Duration(cfg.AuthConfig.JWTAccessExpiry) * time.Hour,
	}
}

func (s *JWTService) GenerateToken(ID int, LastAstiveName int, FirstName string, LastName *string, UserName *string, Description *string, AvatarUrl *string, FullAvatarUrl *string, IsBot bool) (string, error) {
	claims := &Claims{
		ID:             ID,
		FirstName:      FirstName,
		LastName:       getStringValue(LastName),
		UserName:       getStringValue(UserName),
		IsBot:          IsBot,
		LastAstiveName: LastAstiveName,
		Description:    getStringValue(Description),
		AvatarUrl:      getStringValue(AvatarUrl),
		FullAvatarUrl:  getStringValue(FullAvatarUrl),

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "max_app_api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
func (s *JWTService) GenerateTokenPair(ID int, LastAstiveName int, FirstName string, LastName *string, UserName *string, Description *string, AvatarUrl *string, FullAvatarUrl *string, IsBot bool) (accessToken, refreshToken string, err error) {

	accessToken, err = s.GenerateToken(ID, LastAstiveName, FirstName, LastName, UserName, Description, AvatarUrl, FullAvatarUrl, IsBot)
	if err != nil {
		return "", "", err
	}

	raw := uuid.NewString()
	expires := time.Now().Add(s.Expiry * 24)
	signed := fmt.Sprintf("%s.%d", raw, expires.Unix())

	return accessToken, signed, nil
}

func ValidateInitData(initData *dto.WebAppInitData, botToken string) bool {

	dataCheckString := createDataCheckString(initData)

	secretKey := createSecretKey(botToken)

	computedHash := computeSignature(secretKey, dataCheckString)

	return computedHash == initData.Hash
}

func createDataCheckString(initData *dto.WebAppInitData) string {
	params := make(map[string]string)

	params["auth_date"] = strconv.FormatInt(int64(initData.AuthDate), 10)
	params["query_id"] = initData.QueryID

	if initData.StartParam != "" {
		params["start_param"] = initData.StartParam
	}

	userJSON, _ := json.Marshal(initData.User)
	params["user"] = string(userJSON)

	if initData.Chat.ID != 0 {
		chatJSON, _ := json.Marshal(initData.Chat)
		params["chat"] = string(chatJSON)
	}

	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var pairs []string
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, params[key]))
	}

	return strings.Join(pairs, "\n")
}

func createSecretKey(botToken string) []byte {
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	return mac.Sum(nil)
}

func computeSignature(secretKey []byte, dataCheckString string) string {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

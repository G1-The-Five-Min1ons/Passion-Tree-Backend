package jwt

import (
	"errors"
	"os"
	"time"

	"passiontree/internal/auth/model"

	"github.com/golang-jwt/jwt/v5"
)

// JWT configuration constants
const (
	DefaultJWTSecret = "your-secret-key-change-this-in-production"
	AccessTokenTTL   = 24 * time.Hour     // 24 hours
	RefreshTokenTTL  = 7 * 24 * time.Hour // 7 days
)

// getJWTSecret returns JWT secret from environment or default
func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return DefaultJWTSecret
}

// CustomClaims represents JWT claims structure
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Service handles JWT operations
type Service struct {
	secretKey []byte
}

// NewService creates a new JWT service
func NewService() *Service {
	return &Service{
		secretKey: []byte(getJWTSecret()),
	}
}

// GenerateAccessToken generates a new access token
func (s *Service) GenerateAccessToken(user *model.User) (string, error) {
	claims := CustomClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "passion-tree",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken generates a new refresh token
func (s *Service) GenerateRefreshToken(user *model.User) (string, error) {
	claims := CustomClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "passion-tree-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates and parses JWT token
func (s *Service) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractUserID extracts user ID from token
func (s *Service) ExtractUserID(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

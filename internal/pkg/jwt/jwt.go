package jwt

import (
	"errors"
	"os"
	"strconv"
	"time"

	"passiontree/internal/auth/model"

	"github.com/golang-jwt/jwt/v5"
)

// JWT configuration constants
const (
	DefaultJWTSecret     = "your-secret-key-change-this-in-production"
	DefaultAccessTokenTTL  = 24 * time.Hour     // 24 hours
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour // 7 days
)

// Common JWT errors
var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrTokenNotValid     = errors.New("token is not valid yet")
)

// getJWTSecret returns JWT secret from environment or default
func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return DefaultJWTSecret
}

// getAccessTokenTTL returns access token TTL from environment or default
func getAccessTokenTTL() time.Duration {
	if ttlStr := os.Getenv("JWT_ACCESS_TTL"); ttlStr != "" {
		if hours, err := strconv.Atoi(ttlStr); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}
	return DefaultAccessTokenTTL
}

// getRefreshTokenTTL returns refresh token TTL from environment or default
func getRefreshTokenTTL() time.Duration {
	if ttlStr := os.Getenv("JWT_REFRESH_TTL"); ttlStr != "" {
		if hours, err := strconv.Atoi(ttlStr); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}
	return DefaultRefreshTokenTTL
}

// TokenType constants
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// CustomClaims represents JWT claims structure
type CustomClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type,omitempty"` // "access" or "refresh"
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
	now := time.Now()
	claims := CustomClaims{
		UserID:    user.UserID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(getAccessTokenTTL())),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "passion-tree",
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken generates a new refresh token
func (s *Service) GenerateRefreshToken(user *model.User) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    user.UserID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(getRefreshTokenTTL())),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "passion-tree-refresh",
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return s.secretKey, nil
	})

	if err != nil {
		// Check for specific errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValid
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GetTokenExpiration returns the expiration time of a token
func (s *Service) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("token has no expiration")
	}

	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired without validating the signature
func (s *Service) IsTokenExpired(tokenString string) bool {
	expTime, err := s.GetTokenExpiration(tokenString)
	if err != nil {
		return true
	}
	return time.Now().After(expTime)
}

// GenerateTokenPair generates both access and refresh tokens
func (s *Service) GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateRefreshToken validates refresh token specifically
func (s *Service) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Verify it's a refresh token
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}

	// Verify issuer
	if claims.Issuer != "passion-tree-refresh" {
		return nil, errors.New("invalid refresh token issuer")
	}

	return claims, nil
}

// ValidateAccessToken validates access token specifically
func (s *Service) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Verify it's an access token
	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New("not an access token")
	}

	// Verify issuer
	if claims.Issuer != "passion-tree" {
		return nil, errors.New("invalid access token issuer")
	}

	return claims, nil
}

// ExtractUserID extracts user ID from token
func (s *Service) ExtractUserID(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

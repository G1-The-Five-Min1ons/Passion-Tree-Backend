package jwt

import (
	"errors"
	"strconv"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Common JWT errors
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token has expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrTokenNotValid        = errors.New("token is not valid yet")
)

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
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewService creates a new JWT service from config
func NewService(cfg *config.Config) *Service {
	// Parse access token TTL from config (hours)
	accessTTL := 24 * time.Hour
	if hours, err := strconv.Atoi(cfg.JWTAccessTTL); err == nil {
		accessTTL = time.Duration(hours) * time.Hour
	}

	// Parse refresh token TTL from config (hours)
	refreshTTL := 168 * time.Hour // 7 days
	if hours, err := strconv.Atoi(cfg.JWTRefreshTTL); err == nil {
		refreshTTL = time.Duration(hours) * time.Hour
	}

	return &Service{
		secretKey:       []byte(cfg.JWTSecret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateAccessToken generates a new access token
func (s *Service) GenerateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    user.UserID,
		Username:  user.Username,
		Role:      string(user.Role),
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "passion-tree",
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
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
		Role:      string(user.Role),
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
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
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrInvalidToken
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
// This is useful for checking token expiration status even when the token might be invalid
func (s *Service) IsTokenExpired(tokenString string) bool {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claims := &CustomClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil || claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(claims.ExpiresAt.Time)
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

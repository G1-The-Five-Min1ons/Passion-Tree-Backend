package service

import (
	"fmt"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates user and returns token
// identifier can be either username or email
func (s *userServiceImpl) Login(identifier string, password string) (string, error) {
	if identifier == "" {
		return "", apperror.NewBadRequest("username or email is required")
	}
	if password == "" {
		return "", apperror.NewBadRequest("password is required")
	}

	// Try to find user by email first, then by username
	var user *model.User
	var err error

	// Check if identifier is email (contains @)
	if contains := false; len(identifier) > 0 {
		for _, char := range identifier {
			if char == '@' {
				contains = true
				break
			}
		}
		if contains {
			user, err = s.userRepo.GetUserByEmail(identifier)
		} else {
			user, err = s.userRepo.GetUserByUsername(identifier)
		}
	}

	if err != nil {
		return "", apperror.NewInternal(err)
	}
	if user == nil {
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", apperror.NewUnauthorized("invalid username/email or password")
	}

	// Generate JWT token
	jwtService := jwt.NewService()
	token, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", apperror.NewInternal(fmt.Errorf("failed to generate token: %w", err))
	}

	return token, nil
}

// ValidateToken validates JWT token and returns user
func (s *userServiceImpl) ValidateToken(token string) (*model.User, error) {
	if token == "" {
		return nil, apperror.NewBadRequest("token is required")
	}

	// Validate JWT token
	jwtService := jwt.NewService()
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		return nil, apperror.NewUnauthorized("invalid token")
	}

	// Get user from database
	user, _, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user not found")
	}

	return user, nil
}

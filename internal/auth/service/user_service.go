package service

import (
	"database/sql"
	"fmt"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with hashed password
func (s *userServiceImpl) CreateUser(user *model.User, profile *model.Profile) (string, error) {
	if user.Email == "" {
		return "", apperror.NewBadRequest("email is required")
	}
	if user.Password == "" {
		return "", apperror.NewBadRequest("password is required")
	}
	if user.Username == "" {
		return "", apperror.NewBadRequest("username is required")
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetUserByEmail(user.Email)
	if err != nil && err != sql.ErrNoRows {
		return "", apperror.NewInternal(err)
	}
	if existingUser != nil {
		return "", apperror.NewConflict("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternal(fmt.Errorf("failed to hash password: %w", err))
	}
	user.Password = string(hashedPassword)

	// Set default values
	if user.Role == "" {
		user.Role = "user"
	}
	if user.HeartCount == 0 {
		user.HeartCount = 5 // default hearts
	}

	// Set default profile values
	if profile.Level == 0 {
		profile.Level = 1
	}
	if profile.XP == 0 {
		profile.XP = 0
	}
	if profile.LearningStreak == 0 {
		profile.LearningStreak = 0
	}
	if profile.LearningCount == 0 {
		profile.LearningCount = 0
	}
	if profile.HourLearned == 0 {
		profile.HourLearned = 0
	}
	if profile.RankName == "" {
		profile.RankName = "Beginner"
	}

	// Create user and profile
	userID, err := s.userRepo.CreateUser(user, profile)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("user with this email or username already exists")
		}
		return "", apperror.NewInternal(err)
	}

	return userID, nil
}

// GetUserByID retrieves user and profile by ID
func (s *userServiceImpl) GetUserByID(id string) (*model.User, *model.Profile, error) {
	if id == "" {
		return nil, nil, apperror.NewBadRequest("user_id is required")
	}

	user, profile, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, nil, apperror.NewNotFound("user with id '%s' not found", id)
	}

	return user, profile, nil
}

// GetUserByEmail retrieves user by email
func (s *userServiceImpl) GetUserByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, apperror.NewBadRequest("email is required")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user with email '%s' not found", email)
	}

	return user, nil
}

// UpdateUser updates user information
func (s *userServiceImpl) UpdateUser(id string, user *model.User) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	// Check if user exists
	existingUser, _, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if existingUser == nil {
		return apperror.NewNotFound("user with id '%s' not found", id)
	}

	// Hash password if provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return apperror.NewInternal(fmt.Errorf("failed to hash password: %w", err))
		}
		user.Password = string(hashedPassword)
	}

	if err := s.userRepo.UpdateUser(id, user); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("email or username already exists")
		}
		return apperror.NewInternal(err)
	}

	return nil
}

// DeleteUser deletes a user
func (s *userServiceImpl) DeleteUser(id string) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	if err := s.userRepo.DeleteUser(id); err != nil {
		return apperror.NewInternal(err)
	}

	return nil
}

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
		for _, ch := range identifier {
			if ch == '@' {
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

	// TODO: Generate JWT token
	token := fmt.Sprintf("token_%s_%d", user.UserID, time.Now().Unix())

	return token, nil
}

// ValidateToken validates JWT token and returns user
func (s *userServiceImpl) ValidateToken(token string) (*model.User, error) {
	if token == "" {
		return nil, apperror.NewUnauthorized("token is required")
	}

	// TODO: Implement JWT token validation
	return nil, apperror.NewUnauthorized("invalid token")
}

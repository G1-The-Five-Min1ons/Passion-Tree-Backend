package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

type TokenRepository interface {
	CreateToken(token *model.Token) error
	GetTokenByValue(tokenValue string, tokenType string) (*model.Token, error)
	RevokeToken(tokenID string) error
	RevokeTokenByValue(tokenValue string, tokenType string) error
	RevokeAllUserTokens(userID string, tokenType string) error
	IsTokenRevoked(tokenValue string, tokenType string) (bool, error)
	DeleteExpiredTokens() error
	DeleteTokensByUserAndType(userID string, tokenType string) error
}

type tokenRepositoryImpl struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepositoryImpl{
		db: db,
	}
}

// CreateToken creates a new token
func (r *tokenRepositoryImpl) CreateToken(token *model.Token) error {
	ctx := context.Background()

	if token.TokenID == "" {
		token.TokenID = uuid.New().String()
	}

	query := `INSERT INTO Token (token_id, user_id, token, token_type, is_revoke, expire_at) 
	          VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`

	_, err := r.db.ExecContext(ctx, query,
		token.TokenID, token.UserID, token.Token, token.TokenType, token.IsRevoked, token.ExpireAt)
	if err != nil {
		return fmt.Errorf("create token failed: %w", err)
	}

	return nil
}

// GetTokenByValue retrieves a token by its value and type
func (r *tokenRepositoryImpl) GetTokenByValue(tokenValue string, tokenType string) (*model.Token, error) {
	query := `SELECT CONVERT(VARCHAR(36), token_id) as token_id, 
	          CONVERT(VARCHAR(36), user_id) as user_id, 
	          token, token_type, is_revoke, expire_at
	          FROM Token 
	          WHERE token = @p1 AND token_type = @p2 AND is_revoke = 0`

	var token model.Token
	err := r.db.QueryRow(query, tokenValue, tokenType).Scan(
		&token.TokenID, &token.UserID, &token.Token, &token.TokenType,
		&token.IsRevoked, &token.CreatedAt, &token.ExpireAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get token by value failed: %w", err)
	}
	return &token, nil
}

// RevokeToken marks a token as revoked
func (r *tokenRepositoryImpl) RevokeToken(tokenID string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`
	_, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("revoke token failed: %w", err)
	}
	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (r *tokenRepositoryImpl) DeleteExpiredTokens() error {
	query := `DELETE FROM Token WHERE expire_at < GETDATE()`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("delete expired tokens failed: %w", err)
	}
	return nil
}

// DeleteTokensByUserAndType deletes all tokens of a specific type for a user
func (r *tokenRepositoryImpl) DeleteTokensByUserAndType(userID string, tokenType string) error {
	query := `DELETE FROM Token WHERE user_id = @p1 AND token_type = @p2`
	_, err := r.db.Exec(query, userID, tokenType)
	if err != nil {
		return fmt.Errorf("delete tokens by user and type failed: %w", err)
	}
	return nil
}

// RevokeTokenByValue revokes a token by its value
func (r *tokenRepositoryImpl) RevokeTokenByValue(tokenValue string, tokenType string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE token = @p1 AND token_type = @p2`
	result, err := r.db.Exec(query, tokenValue, tokenType)
	if err != nil {
		return fmt.Errorf("revoke token by value failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("token not found or already revoked")
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens of a specific type for a user
func (r *tokenRepositoryImpl) RevokeAllUserTokens(userID string, tokenType string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	_, err := r.db.Exec(query, userID, tokenType)
	if err != nil {
		return fmt.Errorf("revoke all user tokens failed: %w", err)
	}
	return nil
}

// IsTokenRevoked checks if a token is revoked
func (r *tokenRepositoryImpl) IsTokenRevoked(tokenValue string, tokenType string) (bool, error) {
	query := `SELECT is_revoke FROM Token WHERE token = @p1 AND token_type = @p2`

	var isRevoked bool
	err := r.db.QueryRow(query, tokenValue, tokenType).Scan(&isRevoked)
	if err != nil {
		if err == sql.ErrNoRows {
			// Token not found in database - consider it as not revoked
			// (it might be a JWT that was never stored)
			return false, nil
		}
		return false, fmt.Errorf("check token revocation failed: %w", err)
	}

	return isRevoked, nil
}

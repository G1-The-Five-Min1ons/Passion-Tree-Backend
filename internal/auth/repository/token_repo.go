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
	RevokeTokenByValue(tokenValue string, tokenType string) error
	RevokeAllUserTokens(userID string, tokenType string) error
	DeleteExpiredTokens() error
	DeleteTokensByUserAndType(userID string, tokenType string) error
	
	// Token Rotation Methods
	MarkTokenAsRotated(tokenValue string, tokenType string) error
}

type tokenRepositoryImpl struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepositoryImpl{
		db: db,
	}
}

// CreateToken creates a new token with full session tracking
func (r *tokenRepositoryImpl) CreateToken(token *model.Token) error {
	ctx := context.Background()

	if token.TokenID == "" {
		token.TokenID = uuid.New().String()
	}

	query := `INSERT INTO Token (
		token_id, user_id, token, token_type, is_revoke, expire_at,
		device_info, ip_address, user_agent, last_used_at, max_expires_at, parent_token_id, is_rotated
	) VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13)`

	_, err := r.db.ExecContext(ctx, query,
		token.TokenID, token.UserID, token.Token, token.TokenType, token.IsRevoked, token.ExpireAt,
		token.DeviceInfo, token.IPAddress, token.UserAgent, token.LastUsedAt, token.MaxExpiresAt, 
		token.ParentTokenID, token.IsRotated)
	if err != nil {
		return fmt.Errorf("create token failed: %w", err)
	}

	return nil
}

// GetTokenByValue retrieves a token by its value and type with all session tracking fields
func (r *tokenRepositoryImpl) GetTokenByValue(tokenValue string, tokenType string) (*model.Token, error) {
	query := `SELECT 
		CONVERT(VARCHAR(36), token_id) as token_id,
		CONVERT(VARCHAR(36), user_id) as user_id,
		token, token_type, is_revoke, created_at, expire_at,
		device_info, ip_address, user_agent, last_used_at, max_expires_at,
		CONVERT(VARCHAR(36), parent_token_id) as parent_token_id, is_rotated
		FROM Token 
		WHERE token = @p1 AND token_type = @p2`

	var token model.Token
	err := r.db.QueryRow(query, tokenValue, tokenType).Scan(
		&token.TokenID, &token.UserID, &token.Token, &token.TokenType,
		&token.IsRevoked, &token.CreatedAt, &token.ExpireAt,
		&token.DeviceInfo, &token.IPAddress, &token.UserAgent, &token.LastUsedAt, &token.MaxExpiresAt,
		&token.ParentTokenID, &token.IsRotated)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get token by value failed: %w", err)
	}
	return &token, nil
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

// MarkTokenAsRotated marks a token as rotated (replaced by a new token)
func (r *tokenRepositoryImpl) MarkTokenAsRotated(tokenValue string, tokenType string) error {
	query := `UPDATE Token SET is_rotated = 1 WHERE token = @p1 AND token_type = @p2`
	result, err := r.db.Exec(query, tokenValue, tokenType)
	if err != nil {
		return fmt.Errorf("mark token as rotated failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// CreateToken creates a new token
func (r *tokenRepositoryImpl) CreateToken(ctx context.Context, token *model.Token) error {

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
func (r *tokenRepositoryImpl) GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
	query := `SELECT CONVERT(VARCHAR(36), token_id) as token_id, 
	          CONVERT(VARCHAR(36), user_id) as user_id, 
	          token, token_type, is_revoke, expire_at
	          FROM Token 
	          WHERE token = @p1 AND token_type = @p2 AND is_revoke = 0`

	var token model.Token
	err := r.db.QueryRowContext(ctx, query, tokenValue, tokenType).Scan(
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
func (r *tokenRepositoryImpl) RevokeToken(ctx context.Context, tokenID string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("revoke token failed: %w", err)
	}
	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (r *tokenRepositoryImpl) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM Token WHERE expire_at < GETDATE()`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("delete expired tokens failed: %w", err)
	}
	return nil
}

// DeleteTokensByUserAndType deletes all tokens of a specific type for a user
func (r *tokenRepositoryImpl) DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error {
	query := `DELETE FROM Token WHERE user_id = @p1 AND token_type = @p2`
	_, err := r.db.ExecContext(ctx, query, userID, tokenType)
	if err != nil {
		return fmt.Errorf("delete tokens by user and type failed: %w", err)
	}
	return nil
}

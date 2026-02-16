package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/auth/model"

	"github.com/google/uuid"
)

// CreateToken creates a new token
func (r *repositoryImpl) CreateToken(ctx context.Context, token *model.Token) error {

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
func (r *repositoryImpl) GetTokenByValue(ctx context.Context, tokenValue string, tokenType string) (*model.Token, error) {
	query := `SELECT 
		CONVERT(VARCHAR(36), token_id) as token_id,
		CONVERT(VARCHAR(36), user_id) as user_id,
		token, token_type, is_revoke, created_at, expire_at,
		device_info, ip_address, user_agent, last_used_at, max_expires_at,
		CONVERT(VARCHAR(36), parent_token_id) as parent_token_id, is_rotated
		FROM Token 
		WHERE token = @p1 AND token_type = @p2 AND is_revoke = 0`

	var token model.Token
	err := r.db.QueryRowContext(ctx, query, tokenValue, tokenType).Scan(
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

// RevokeToken marks a token as revoked
func (r *repositoryImpl) RevokeToken(ctx context.Context, tokenID string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE token_id = @p1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("revoke token failed: %w", err)
	}
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user (invalidates all sessions)
func (r *repositoryImpl) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	_, err := r.db.ExecContext(ctx, query, userID, model.TokenTypeRefresh)
	if err != nil {
		return fmt.Errorf("revoke all user refresh tokens failed: %w", err)
	}
	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (r *repositoryImpl) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM Token WHERE expire_at < GETDATE()`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("delete expired tokens failed: %w", err)
	}
	return nil
}

// DeleteTokensByUserAndType deletes all tokens of a specific type for a user
func (r *repositoryImpl) DeleteTokensByUserAndType(ctx context.Context, userID string, tokenType string) error {
	query := `DELETE FROM Token WHERE user_id = @p1 AND token_type = @p2`
	_, err := r.db.ExecContext(ctx, query, userID, tokenType)
	if err != nil {
		return fmt.Errorf("delete tokens by user and type failed: %w", err)
	}
	return nil
}

// RevokeTokenByValue revokes a token by its value
func (r *repositoryImpl) RevokeTokenByValue(ctx context.Context, tokenValue string, tokenType string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE token = @p1 AND token_type = @p2`
	result, err := r.db.ExecContext(ctx, query, tokenValue, tokenType)
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
func (r *repositoryImpl) RevokeAllUserTokens(ctx context.Context, userID string, tokenType string) error {
	query := `UPDATE Token SET is_revoke = 1 WHERE user_id = @p1 AND token_type = @p2 AND is_revoke = 0`
	_, err := r.db.ExecContext(ctx, query, userID, tokenType)
	if err != nil {
		return fmt.Errorf("revoke all user tokens failed: %w", err)
	}
	return nil
}

// MarkTokenAsRotated marks a token as rotated (replaced by a new token)
func (r *repositoryImpl) MarkTokenAsRotated(ctx context.Context, tokenValue string, tokenType string) error {
	query := `UPDATE Token SET is_rotated = 1 WHERE token = @p1 AND token_type = @p2`
	result, err := r.db.ExecContext(ctx, query, tokenValue, tokenType)
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

// GetActiveUserSessions retrieves all active (non-revoked, non-rotated, non-expired) sessions for a user
func (r *repositoryImpl) GetActiveUserSessions(ctx context.Context, userID string, tokenType string) ([]*model.Token, error) {
	query := `SELECT 
		CONVERT(VARCHAR(36), token_id) as token_id,
		CONVERT(VARCHAR(36), user_id) as user_id,
		token, token_type, is_revoke, created_at, expire_at,
		device_info, ip_address, user_agent, last_used_at, max_expires_at,
		CONVERT(VARCHAR(36), parent_token_id) as parent_token_id, is_rotated
	FROM Token 
	WHERE user_id = @p1 
		AND token_type = @p2 
		AND is_revoke = 0 
		AND is_rotated = 0
		AND expire_at > GETDATE()
	ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID, tokenType)
	if err != nil {
		return nil, fmt.Errorf("get active user sessions failed: %w", err)
	}
	defer rows.Close()

	var sessions []*model.Token
	for rows.Next() {
		var session model.Token
		err := rows.Scan(
			&session.TokenID, &session.UserID, &session.Token, &session.TokenType,
			&session.IsRevoked, &session.CreatedAt, &session.ExpireAt,
			&session.DeviceInfo, &session.IPAddress, &session.UserAgent, 
			&session.LastUsedAt, &session.MaxExpiresAt,
			&session.ParentTokenID, &session.IsRotated,
		)
		if err != nil {
			return nil, fmt.Errorf("scan session row failed: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return sessions, nil
}

// RevokeTokenByIDForUser revokes a token only if it belongs to the specified user (for security)
func (r *repositoryImpl) RevokeTokenByIDForUser(ctx context.Context, tokenID string, userID string) error {
	query := `UPDATE Token 
		SET is_revoke = 1 
		WHERE token_id = @p1 
			AND user_id = @p2 
			AND is_revoke = 0 
			AND is_rotated = 0`
	result, err := r.db.ExecContext(ctx, query, tokenID, userID)
	if err != nil {
		return fmt.Errorf("revoke token by id for user failed: %w", err)
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


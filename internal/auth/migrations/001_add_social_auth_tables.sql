-- Migration: Add Social Authentication Tables
-- Description: Creates tables for managing social authentication with Discord and Google
-- Date: 2026-02-13

-- =============================================
-- Table: social_auth_providers
-- Description: Stores user connections to social auth providers
-- =============================================
CREATE TABLE social_auth_providers (
    id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    user_id UNIQUEIDENTIFIER NOT NULL,
    provider VARCHAR(50) NOT NULL, -- 'google', 'discord'
    provider_user_id VARCHAR(255) NOT NULL, -- User ID from the provider
    provider_email VARCHAR(255) NULL, -- Email from the provider
    provider_username VARCHAR(255) NULL, -- Username from the provider (mainly for Discord)
    access_token NVARCHAR(MAX) NULL, -- Encrypted access token
    refresh_token NVARCHAR(MAX) NULL, -- Encrypted refresh token
    token_expires_at DATETIME2 NULL, -- When the access token expires
    provider_data NVARCHAR(MAX) NULL, -- JSON string of additional provider data
    is_active BIT NOT NULL DEFAULT 1, -- Whether this connection is active
    create_at DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
    update_at DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
    
    -- Foreign key to users table
    CONSTRAINT FK_social_auth_providers_users FOREIGN KEY (user_id) REFERENCES dbo.Users(user_id) ON DELETE CASCADE,
    
    -- Ensure one provider account can only be linked to one user
    CONSTRAINT UQ_provider_user UNIQUE (provider, provider_user_id),
    
    -- Index for quick lookups
    INDEX IX_social_auth_providers_user_id (user_id),
    INDEX IX_social_auth_providers_provider (provider),
    INDEX IX_social_auth_providers_provider_email (provider_email)
);

-- =============================================
-- Table: social_auth_states
-- Description: Stores OAuth state tokens for CSRF protection
-- =============================================
CREATE TABLE social_auth_states (
    state_token VARCHAR(255) PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    redirect_uri VARCHAR(500) NULL,
    create_at DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
    expires_at DATETIME2 NOT NULL,
    used BIT NOT NULL DEFAULT 0,
    
    -- Index for cleanup of expired states
    INDEX IX_social_auth_states_expires_at (expires_at)
);

-- =============================================
-- Trigger: Update timestamp on social_auth_providers
-- =============================================
GO
CREATE TRIGGER trg_social_auth_providers_update
ON social_auth_providers
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    UPDATE social_auth_providers
    SET update_at = GETUTCDATE()
    FROM social_auth_providers sap
    INNER JOIN inserted i ON sap.id = i.id;
END;
GO

-- =============================================
-- Stored Procedure: Clean up expired OAuth states
-- =============================================
CREATE PROCEDURE sp_cleanup_expired_oauth_states
AS
BEGIN
    SET NOCOUNT ON;
    DELETE FROM social_auth_states
    WHERE expires_at < GETUTCDATE() OR (used = 1 AND create_at < DATEADD(HOUR, -1, GETUTCDATE()));
END;
GO

-- =============================================
-- Comments for documentation
-- =============================================
EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Stores social authentication provider connections for users',
    @level0type = N'SCHEMA', @level0name = N'dbo',
    @level1type = N'TABLE',  @level1name = N'social_auth_providers';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Stores OAuth state tokens for CSRF protection during social auth flow',
    @level0type = N'SCHEMA', @level0name = N'dbo',
    @level1type = N'TABLE',  @level1name = N'social_auth_states';

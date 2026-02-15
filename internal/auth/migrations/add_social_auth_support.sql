-- Migration: Add Social Auth Support to Users Table
-- Description: Adds auth_provider and provider_user_id columns for OAuth integration
-- Date: 2026-02-13

-- Add social auth columns to users table
ALTER TABLE users
ADD 
    auth_provider VARCHAR(20) DEFAULT 'local' NOT NULL,
    provider_user_id VARCHAR(255) NULL;

-- Create index for faster social auth lookups
CREATE INDEX idx_users_provider_lookup 
ON users(auth_provider, provider_user_id);

-- Make password nullable for social auth users
ALTER TABLE users
ALTER COLUMN password VARCHAR(255) NULL;

-- Update existing users to have 'local' auth provider
UPDATE users
SET auth_provider = 'local'
WHERE auth_provider IS NULL;

-- Add constraint to ensure provider_user_id is unique per provider
CREATE UNIQUE INDEX idx_users_provider_unique
ON users(auth_provider, provider_user_id)
WHERE provider_user_id IS NOT NULL;

-- Comments for documentation
EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Authentication provider: local, google, discord', 
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users',
    @level2type = N'COLUMN', @level2name = 'auth_provider';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'User ID from OAuth provider', 
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users',
    @level2type = N'COLUMN', @level2name = 'provider_user_id';

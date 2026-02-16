-- Migration: Add require_2fa_next_login flag to Users table
-- Purpose: Security enhancement - require 2FA verification after token theft detection

-- Add require_2fa_next_login column to Users table
ALTER TABLE Users
ADD require_2fa_next_login BIT NOT NULL DEFAULT 0;

-- Add index for faster queries filtering users requiring 2FA
CREATE NONCLUSTERED INDEX IX_Users_Require2FA 
ON Users(require_2fa_next_login)
WHERE require_2fa_next_login = 1;

-- Add comment explaining the column
EXEC sp_addextendedproperty 
    @name = N'MS_Description',
    @value = N'Security flag: When set to 1, user must verify identity (2FA) on next login. Automatically set when token theft is detected.',
    @level0type = N'SCHEMA', @level0name = N'dbo',
    @level1type = N'TABLE',  @level1name = N'Users',
    @level2type = N'COLUMN', @level2name = N'require_2fa_next_login';

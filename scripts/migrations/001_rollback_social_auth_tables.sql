-- Rollback Migration: Remove Social Authentication Tables
-- Description: Drops all tables and objects created for social authentication
-- Date: 2026-02-13

-- Drop stored procedures
IF OBJECT_ID('sp_cleanup_expired_oauth_states', 'P') IS NOT NULL
    DROP PROCEDURE sp_cleanup_expired_oauth_states;
GO

-- Drop triggers
IF OBJECT_ID('trg_social_auth_providers_update', 'TR') IS NOT NULL
    DROP TRIGGER trg_social_auth_providers_update;
GO

-- Drop tables (in reverse order of dependencies)
IF OBJECT_ID('dbo.social_auth_states', 'U') IS NOT NULL
    DROP TABLE dbo.social_auth_states;

IF OBJECT_ID('dbo.social_auth_providers', 'U') IS NOT NULL
    DROP TABLE dbo.social_auth_providers;

PRINT 'Social authentication tables successfully removed';

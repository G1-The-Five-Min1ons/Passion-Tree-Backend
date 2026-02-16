# Database Migrations - Authentication Module

This directory contains SQL migration scripts for authentication-related database changes.

## Migration Files

### 001_add_social_auth_tables.sql
Creates the necessary tables for social authentication (Discord and Google):

**Tables Created:**
- `social_auth_providers`: Stores user connections to social auth providers
  - Links users to their Google/Discord accounts
  - Stores provider user ID, email, and tokens
  - Tracks token expiration
  - Stores additional provider data as JSON

- `social_auth_states`: Stores OAuth state tokens for CSRF protection
  - Temporary storage for OAuth flow state
  - Includes expiration and usage tracking
  - Prevents CSRF attacks during OAuth flow

**Additional Objects:**
- Trigger: `trg_social_auth_providers_update` - Auto-updates `updated_at` timestamp
- Stored Procedure: `sp_cleanup_expired_oauth_states` - Cleans up expired OAuth states

### 002_add_require_2fa_flag.sql
Adds security enhancement for token theft detection:

**Changes:**
- Adds `require_2fa_next_login` column to Users table
  - Type: BIT (boolean)
  - Default: 0 (false)
  - Purpose: Flag users requiring additional security verification after token theft detection

- Creates index `IX_Users_Require2FA` for performance
- Adds column description via extended properties

**Security Flow:**
1. When token theft/reuse is detected, this flag is set to true
2. User must verify identity (2FA) on next login attempt
3. Flag is cleared after successful verification

### Rollback Scripts
- `001_rollback_social_auth_tables.sql` - Removes social auth tables
- `002_rollback_require_2fa_flag.sql` - Removes 2FA flag column

## How to Run Migrations

### Apply Migration
```powershell
# Using sqlcmd
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 001_add_social_auth_tables.sql
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 002_add_require_2fa_flag.sql

# Using Azure Data Studio or SQL Server Management Studio
# Simply open the file and execute
```

### Rollback Migration
```powershell
# Rollback in reverse order
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 002_rollback_require_2fa_flag.sql
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 001_rollback_social_auth_tables.sql
```

## Migration Order
Migrations should be applied in numerical order:
1. 001_add_social_auth_tables.sql
2. 002_add_require_2fa_flag.sql

## Notes
- Always review migrations before applying to production
- Keep rollback scripts synchronized with forward migrations
- Test migrations in development/staging environment first
- Always backup database before running migrations in production

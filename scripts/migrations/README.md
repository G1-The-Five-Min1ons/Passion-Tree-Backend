# Database Migrations

This directory contains SQL migration scripts for the Passion Tree Backend database.

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

### 001_rollback_social_auth_tables.sql
Rollback script to remove all social authentication tables and related objects.

## How to Run Migrations

### Apply Migration
```powershell
# Using sqlcmd
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 001_add_social_auth_tables.sql

# Using Azure Data Studio or SQL Server Management Studio
# Simply open the file and execute
```

### Rollback Migration
```powershell
sqlcmd -S your_server -d your_database -U your_user -P your_password -i 001_rollback_social_auth_tables.sql
```

## Migration Order
Migrations should be applied in numerical order:
1. 001_add_social_auth_tables.sql

## Notes
- Always review migrations before applying to production
- Take a database backup before running migrations
- Test migrations in a development environment first
- The social_auth_providers table has a CASCADE DELETE on user_id - deleting a user will remove their social auth connections
- OAuth states should be cleaned up periodically using the `sp_cleanup_expired_oauth_states` stored procedure

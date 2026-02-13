# Social Authentication Implementation Guide

## Overview
This implementation adds Google and Discord OAuth2 authentication to the Passion Tree backend using the official Go OAuth2 SDK.

## Architecture

### File Structure
```
internal/auth/
├── handler/
│   └── social_auth_handler.go    # HTTP handlers for OAuth flows
├── model/
│   └── user_model.go              # Updated with social auth fields
├── repository/
│   └── social_auth_repo.go        # Database operations for social auth
├── service/
│   └── social_auth_service.go     # Business logic for OAuth
└── routes.go                      # Updated routes with social auth endpoints
```

## Setup Instructions

### 1. Install Required Dependencies

Run the following commands in the `Passion-Tree-Backend` directory:

```bash
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google
```

### 2. Database Migration

Execute the migration script to add social auth support:

```bash
# Connect to your Azure SQL Database and run:
sqlcmd -S your-server.database.windows.net -d your-database -U your-user -P your-password -i scripts/migrations/add_social_auth_support.sql
```

Or use Azure Data Studio / SSMS to execute:
- File: `scripts/migrations/add_social_auth_support.sql`

### 3. Environment Variables

Add the following environment variables to your `.env` file:

```env
# Google OAuth2
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:5000/auth/google/callback

# Discord OAuth2
DISCORD_CLIENT_ID=your-discord-client-id
DISCORD_CLIENT_SECRET=your-discord-client-secret
DISCORD_REDIRECT_URL=http://localhost:5000/auth/discord/callback

# App URL (for redirects)
APP_URL=http://localhost:5000
```

### 4. OAuth Provider Setup

#### Google OAuth2 Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Configure the OAuth consent screen
6. Add authorized redirect URIs:
   - Development: `http://localhost:5000/auth/google/callback`
   - Production: `https://your-domain.com/auth/google/callback`
7. Copy the Client ID and Client Secret to your `.env` file

#### Discord OAuth2 Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to "OAuth2" section
4. Add redirect URIs:
   - Development: `http://localhost:5000/auth/discord/callback`
   - Production: `https://your-domain.com/auth/discord/callback`
5. Copy the Client ID and Client Secret to your `.env` file
6. Under "OAuth2" → "Scopes", ensure `identify` and `email` are available

## API Endpoints

### Google Authentication

#### 1. Initiate Google Login
```http
GET /auth/google
```

**Response:**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

**Usage:**
1. Frontend calls this endpoint
2. Redirect user to the `auth_url`
3. User authenticates with Google
4. Google redirects to callback URL

#### 2. Google Callback
```http
GET /auth/google/callback?code=xxx&state=xxx
```

**Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "user_id": "uuid",
    "username": "john.doe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "student"
  }
}
```

### Discord Authentication

#### 1. Initiate Discord Login
```http
GET /auth/discord
```

**Response:**
```json
{
  "auth_url": "https://discord.com/api/oauth2/authorize?..."
}
```

#### 2. Discord Callback
```http
GET /auth/discord/callback?code=xxx&state=xxx
```

**Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "user_id": "uuid",
    "username": "discordUser123",
    "email": "user@example.com",
    "first_name": "Discord",
    "last_name": "User",
    "role": "student"
  }
}
```

## Frontend Integration Example

### React/TypeScript Example

```typescript
// Google Login
const handleGoogleLogin = async () => {
  try {
    const response = await fetch('http://localhost:5000/auth/google');
    const data = await response.json();
    
    // Redirect to Google OAuth
    window.location.href = data.auth_url;
  } catch (error) {
    console.error('Login failed:', error);
  }
};

// Handle callback (on callback page)
const handleOAuthCallback = async () => {
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');
  
  if (token) {
    // Store token
    localStorage.setItem('authToken', token);
    
    // Redirect to dashboard
    window.location.href = '/dashboard';
  }
};

// Discord Login
const handleDiscordLogin = async () => {
  try {
    const response = await fetch('http://localhost:5000/auth/discord');
    const data = await response.json();
    
    // Redirect to Discord OAuth
    window.location.href = data.auth_url;
  } catch (error) {
    console.error('Login failed:', error);
  }
};
```

### Flutter Example

```dart
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:url_launcher/url_launcher.dart';

// Google Login
Future<void> loginWithGoogle() async {
  final response = await http.get(
    Uri.parse('http://localhost:5000/auth/google'),
  );
  
  if (response.statusCode == 200) {
    final data = json.decode(response.body);
    final authUrl = data['auth_url'];
    
    // Open browser for OAuth
    if (await canLaunch(authUrl)) {
      await launch(authUrl);
    }
  }
}

// Discord Login
Future<void> loginWithDiscord() async {
  final response = await http.get(
    Uri.parse('http://localhost:5000/auth/discord'),
  );
  
  if (response.statusCode == 200) {
    final data = json.decode(response.body);
    final authUrl = data['auth_url'];
    
    if (await canLaunch(authUrl)) {
      await launch(authUrl);
    }
  }
}
```

## Security Features

1. **CSRF Protection**: State parameter validation
2. **Secure Cookies**: HTTPOnly cookies for state storage
3. **Email Verification**: Social auth users are auto-verified
4. **Account Linking**: Automatic linking if email already exists
5. **JWT Tokens**: Secure token generation with expiration

## Database Schema Changes

The migration adds the following columns to the `users` table:

```sql
auth_provider VARCHAR(20) DEFAULT 'local' NOT NULL
provider_user_id VARCHAR(255) NULL
```

- `auth_provider`: Identifies the authentication method (local, google, discord)
- `provider_user_id`: Stores the user ID from the OAuth provider
- Password field is now nullable for social auth users

## User Flow

### New User (Social Auth)
1. User clicks "Login with Google/Discord"
2. Frontend fetches auth URL from backend
3. User is redirected to OAuth provider
4. User authorizes the app
5. Provider redirects to callback with code
6. Backend exchanges code for access token
7. Backend fetches user info from provider
8. Backend creates new user account
9. Backend generates JWT token
10. User is logged in

### Existing User (Same Email)
1. User performs social auth
2. Backend finds existing user by email
3. Backend links social provider to existing account
4. User is logged in with existing account

### Returning Social Auth User
1. User performs social auth
2. Backend finds user by provider + provider_user_id
3. Backend generates new JWT token
4. User is logged in

## Testing

### Manual Testing with cURL

```bash
# 1. Get Google auth URL
curl -X GET http://localhost:5000/auth/google

# 2. Get Discord auth URL
curl -X GET http://localhost:5000/auth/discord

# Note: Callbacks must be tested through a browser due to OAuth flow
```

### Postman Testing

1. Create a new GET request to `/auth/google` or `/auth/discord`
2. Copy the `auth_url` from the response
3. Open the URL in a browser
4. Complete the OAuth flow
5. Observe the callback response

## Production Considerations

1. **HTTPS Only**: Enable secure cookies in production
   ```go
   Secure: true, // In social_auth_handler.go
   ```

2. **Environment-specific Redirects**: Update redirect URLs for production
   ```env
   GOOGLE_REDIRECT_URL=https://your-domain.com/auth/google/callback
   DISCORD_REDIRECT_URL=https://your-domain.com/auth/discord/callback
   ```

3. **JWT Secret**: Use a strong, randomly generated secret
   - Currently hardcoded in service (TODO: Move to environment variable)

4. **Rate Limiting**: Consider adding rate limiting to OAuth endpoints

5. **Logging**: Monitor OAuth failures and suspicious activities

## Troubleshooting

### Common Issues

1. **"Invalid state parameter" error**
   - Ensure cookies are enabled
   - Check cookie domain settings
   - Verify CSRF protection is working

2. **"Failed to exchange code" error**
   - Verify client ID and secret are correct
   - Check redirect URI matches exactly
   - Ensure OAuth app is properly configured

3. **"Email already exists" with different provider**
   - This is expected behavior
   - Account will be automatically linked
   - User can use either auth method

4. **Database errors**
   - Ensure migration has been run
   - Check database connection
   - Verify user table schema

## Next Steps

1. Add frontend OAuth integration
2. Implement account unlinking
3. Add profile picture sync from OAuth providers
4. Implement refresh token support
5. Add more OAuth providers (GitHub, Microsoft, etc.)
6. Move JWT secret to environment variable
7. Add comprehensive unit tests
8. Add integration tests

## Support

For issues or questions:
- Check logs for detailed error messages
- Review OAuth provider documentation
- Verify environment variables are set correctly
- Ensure database migration was successful

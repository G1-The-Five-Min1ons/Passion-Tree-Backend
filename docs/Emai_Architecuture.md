User Registration Flow:
1. Register → CreateUser()
2. Generate verification token
3. Save token to Token table (token_type='email_verification')
4. Send email with token
5. User clicks link → VerifyEmail()
6. Check token from Token table
7. Update user.is_email_verified = true
8. Revoke token
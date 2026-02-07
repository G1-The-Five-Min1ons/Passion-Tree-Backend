# Email Verification Feature

## Overview
ฟีเจอร์การยืนยันอีเมลช่วยให้มั่นใจว่าผู้ใช้ใช้อีเมลจริงในการลงทะเบียน โดยใช้ Token table เพื่อจัดการ verification tokens

## Features
- ✅ ส่งอีเมลยืนยันอัตโนมัติหลังจากลงทะเบียน
- ✅ ลิงก์ยืนยันหมดอายุใน 24 ชั่วโมง
- ✅ สามารถส่งอีเมลยืนยันใหม่ได้
- ✅ ตรวจสอบสถานะการยืนยันอีเมล
- ✅ ใช้ Token table สำหรับจัดการ tokens ทุกประเภท

## Architecture Overview

### การออกแบบ
ระบบใช้ **Token table** เป็นศูนย์กลางในการจัดการ tokens ทุกประเภท ไม่ว่าจะเป็น:
- Email verification tokens
- JWT refresh tokens  
- Password reset tokens (future)

### ข้อดี
1. **Scalability** - ขยายได้ง่าย เพิ่ม token type ใหม่ได้โดยไม่ต้องแก้ schema
2. **Maintainability** - จัดการ tokens ได้จากที่เดียว
3. **Security** - ควบคุม token lifecycle ได้ดี (create, validate, revoke, cleanup)
4. **Performance** - มี indexes สำหรับ query ที่เร็ว

### Layer Structure
```
Handler → Service → Repository → Database
   ↓         ↓          ↓
 HTTP    Business    Data Access
Layer     Logic       Layer
```

**Repositories:**
- `UserRepository` - จัดการข้อมูล users
- `TokenRepository` - จัดการข้อมูล tokens

**Services:**
- `UserService` - business logic สำหรับ users และ email verification
- `EmailService` - ส่งอีเมล verification

## Database Schema Updates
เพิ่ม columns ใหม่ในตาราง `users` และ `Token`:

### User Table
```sql
-- เพิ่ม column นี้ในตาราง users
ALTER TABLE [user] ADD 
    is_email_verified BIT NOT NULL DEFAULT 0;
```

### Token Table
```sql
-- เพิ่ม column นี้ในตาราง Token
ALTER TABLE Token ADD 
    token_type VARCHAR(50) NOT NULL DEFAULT 'refresh_token';

-- เพิ่ม indexes สำหรับ performance
CREATE INDEX IX_Token_token_type ON Token(token_type);
CREATE INDEX IX_Token_user_token_type ON Token(user_id, token_type);
```

## Token Types
ระบบรองรับ token หลายประเภท:
- `refresh_token` - สำหรับ JWT refresh
- `email_verification` - สำหรับยืนยันอีเมล
- `password_reset` - สำหรับรีเซ็ตรหัสผ่าน (future)

## Environment Configuration
อัพเดต `.env` ใน `Passion-Tree-Infrastructure` folder:

```env
# SMTP Configuration for Email Verification
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=passiontree.noreply@gmail.com
SMTP_PASSWORD=<Google App Password>
SMTP_FROM_EMAIL=noreply@passiontree.com

# Frontend URL สำหรับ verification link
# Development:
APP_URL=http://localhost:5173

# Production:
APP_URL=https://passiontree.azurewebsites.net
```

### Gmail Configuration
1. ไปที่ Google Account Settings
2. เปิด 2-Factor Authentication
3. สร้าง App Password ที่ https://myaccount.google.com/apppasswords
4. ใช้ App Password ใน `SMTP_PASSWORD`

### Other Email Providers
- **SendGrid**: SMTP_HOST=smtp.sendgrid.net, SMTP_PORT=587
- **AWS SES**: SMTP_HOST=email-smtp.us-east-1.amazonaws.com, SMTP_PORT=587
- **Mailgun**: SMTP_HOST=smtp.mailgun.org, SMTP_PORT=587

## API Endpoints

### 1. Register (Modified)
```http
POST /auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user_id": "uuid",
    "token": "jwt-token"
  }
}
```

**Note:** อีเมลยืนยันจะถูกส่งอัตโนมัติหลังลงทะเบียนสำเร็จ

### 2. Verify Email
```http
GET /auth/verify-email?token=verification-token
```

**Success Response:**
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "invalid or expired verification token"
}
```

### 3. Resend Verification Email
```http
POST /auth/resend-verification
Content-Type: application/json

{
  "email": "john@example.com"
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Verification email sent successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "email is already verified"
}
```

## Data Models

### User Model
User model มีฟิลด์สำหรับเก็บสถานะการยืนยันอีเมล:

```go
type User struct {
    UserID          string    `json:"user_id"`
    Username        string    `json:"username"`
    Email           string    `json:"email"`
    Password        string    `json:"-"`
    FirstName       string    `json:"first_name"`
    LastName        string    `json:"last_name"`
    Role            string    `json:"role"`
    HeartCount      int       `json:"heart_count"`
    IsEmailVerified bool      `json:"is_email_verified"`  // NEW - สถานะการยืนยันอีเมล
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Token Model
Token model ใช้สำหรับเก็บ verification tokens และ tokens ประเภทอื่นๆ:

```go
type Token struct {
    TokenID   string    `json:"token_id"`
    UserID    string    `json:"user_id"`
    Token     string    `json:"token"`
    TokenType string    `json:"token_type"`  // NEW - ประเภทของ token
    IsRevoked bool      `json:"is_revoked"`
    CreatedAt time.Time `json:"created_at"`
    ExpireAt  time.Time `json:"expire_at"`
}
```

**Token Types:**
- `email_verification` - สำหรับยืนยันอีเมล
- `refresh_token` - สำหรับ JWT refresh
- `password_reset` - สำหรับรีเซ็ตรหัสผ่าน (future)

## Email Template
อีเมลยืนยันจะมีรูปแบบ HTML พร้อม:
- หัวข้อ: "Email Verification - Passion Tree"
- ปุ่มยืนยันอีเมล
- ลิงก์สำรอง (กรณีปุ่มไม่ทำงาน)
- ข้อความเตือนว่าลิงก์หมดอายุใน 24 ชั่วโมง

## Security Considerations
- ✅ Verification token สร้างด้วย crypto/rand (32 bytes)
- ✅ Token เก็บเป็น hexadecimal string (64 characters)
- ✅ Token หมดอายุภายใน 24 ชั่วโมง
- ✅ Token ถูก revoke หลังยืนยันสำเร็จ
- ✅ ตรวจสอบ expiry และ revoked status ก่อนยืนยัน
- ✅ ใช้ HTTPS สำหรับ verification links (production)
- ✅ Rate limiting สำหรับ resend verification (แนะนำ)

## Implementation Details

### Key Files
```
internal/auth/
├── model/
│   ├── user_model.go          # User struct with IsEmailVerified
│   └── token_model.go         # Token struct with TokenType
├── repository/
│   ├── user_repo.go           # User CRUD + UpdateEmailVerified
│   └── token_repo.go          # Token CRUD + cleanup methods
├── service/
│   ├── user_service.go        # VerifyEmail, ResendVerification logic
│   └── email_service.go       # SMTP email sending
├── handler/
│   └── user_hanler.go         # HTTP handlers
└── routes.go                  # Route registration
```

### Important Functions

**TokenRepository:**
- `CreateToken()` - สร้าง token ใหม่
- `GetTokenByValue()` - ค้นหา token ตาม value และ type
- `RevokeToken()` - revoke token
- `DeleteTokensByUserAndType()` - ลบ tokens เก่า
- `DeleteExpiredTokens()` - cleanup tokens หมดอายุ

**UserService:**
- `CreateUser()` - สร้าง user + token + ส่งอีเมล
- `VerifyEmail()` - ตรวจสอบ token และอัพเดตสถานะ
- `ResendVerificationEmail()` - ส่งอีเมลใหม่

**EmailService:**
- `SendVerificationEmail()` - ส่งอีเมล HTML
- `GenerateVerificationToken()` - สร้าง random token
- `GetVerificationTokenExpiry()` - คำนวณเวลาหมดอายุ

## Flow Diagram

### Registration & Email Verification Flow
```
┌─────────────────────────────────────────────────────────────┐
│ 1. User Registration                                        │
│    POST /auth/register                                      │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Create User Account                                      │
│    - Hash password                                          │
│    - Set is_email_verified = false                          │
│    - Insert into users table                                │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Generate Verification Token                              │
│    - Create random token (64 chars)                         │
│    - Set expiry = now + 24 hours                            │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Save Token to Database                                   │
│    INSERT INTO Token:                                       │
│    - user_id, token                                         │
│    - token_type = 'email_verification'                      │
│    - expire_at, is_revoke = false                           │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. Send Verification Email                                  │
│    - To: user's email                                       │
│    - Link: {APP_URL}/auth/verify-email?token=xxx            │
│    - HTML template with button                              │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. User Clicks Verification Link                            │
│    GET /auth/verify-email?token=xxx                         │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. Validate Token                                           │
│    - Query Token table by token value                       │
│    - Check token_type = 'email_verification'                │
│    - Check is_revoke = false                                │
│    - Check expire_at > now                                  │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ 8. Update User & Token                                      │
│    - UPDATE users SET is_email_verified = true              │
│    - UPDATE Token SET is_revoke = true                      │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│ ✓ Email Verified Successfully                               │
└─────────────────────────────────────────────────────────────┘
```

### Resend Verification Flow
```
User Request → Check if verified → Delete old tokens → 
Generate new token → Save to Token table → Send email
```

## Testing

### Prerequisites
- ตั้งค่า SMTP credentials ใน `.env`
- Database มี tables และ columns ครบถ้วน
- Backend server รันที่ `http://localhost:5000`

### 1. Test Registration & Email Sending
```bash
curl -X POST http://localhost:5000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Expected:**
- Status: 201 Created
- Response contains `user_id` and `token`
- Email sent to test@example.com
- Check Token table: new record with `token_type='email_verification'`
- Check users table: `is_email_verified=false`

### 2. Test Email Verification
```bash
# Copy token from email or database
TOKEN="your-verification-token-here"

curl "http://localhost:5000/auth/verify-email?token=$TOKEN"
```

**Expected:**
- Status: 200 OK
- Message: "Email verified successfully"
- Check users table: `is_email_verified=true`
- Check Token table: token `is_revoke=true`

### 3. Test Resend Verification
```bash
curl -X POST http://localhost:5000/auth/resend-verification \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
```

**Expected:**
- Status: 200 OK
- Message: "Verification email sent successfully"
- Old tokens deleted from Token table
- New token created with new expiry

### 4. Test Edge Cases

#### Already Verified Email
```bash
# Verify อีกครั้งหลังจาก verified แล้ว
curl "http://localhost:5000/auth/verify-email?token=$TOKEN"
```
**Expected:** 400 Bad Request - "email is already verified"

#### Expired Token
```bash
# ใช้ token ที่หมดอายุ (>24 hours)
curl "http://localhost:5000/auth/verify-email?token=expired-token"
```
**Expected:** 400 Bad Request - "verification token has expired"

#### Invalid Token
```bash
# ใช้ token ที่ไม่มีในระบบ
curl "http://localhost:5000/auth/verify-email?token=invalid-token"
```
**Expected:** 400 Bad Request - "invalid verification token"

### 5. Verify Database State
```sql
-- ตรวจสอบ user
SELECT user_id, email, is_email_verified 
FROM users 
WHERE email = 'test@example.com';

-- ตรวจสอบ tokens
SELECT token_id, token_type, is_revoke, expire_at
FROM Token
WHERE user_id = 'your-user-id' 
  AND token_type = 'email_verification';
```

## Troubleshooting

### Email Not Sending

**Symptoms:**
- Registration สำเร็จแต่ไม่ได้รับอีเมล
- Console แสดง warning: "failed to send verification email"

**Solutions:**
1. ✅ ตรวจสอบ SMTP credentials ใน `.env`
   ```bash
   # ตรวจสอบว่าค่าถูกต้อง
   echo $SMTP_HOST
   echo $SMTP_USERNAME
   ```

2. ✅ ตรวจสอบว่าใช้ **App Password** สำหรับ Gmail (ไม่ใช่รหัสผ่านปกติ)
   - ไปที่: https://myaccount.google.com/apppasswords
   - สร้าง App Password ใหม่
   - อัพเดตใน SMTP_PASSWORD

3. ✅ ตรวจสอบ Firewall/Network
   ```bash
   # ทดสอบ connection ไปยัง SMTP server
   telnet smtp.gmail.com 587
   ```

4. ✅ ตรวจสอบ logs ใน console
   ```bash
   # ดู error details
   docker logs backend-container
   # หรือ
   ./bin/app
   ```

### Token Expired

**Symptoms:**
- คลิกลิงก์ แต่ได้ข้อความ "verification token has expired"

**Solutions:**
1. ✅ ใช้ Resend Verification API
   ```bash
   curl -X POST http://localhost:5000/auth/resend-verification \
     -H "Content-Type: application/json" \
     -d '{"email": "your-email@example.com"}'
   ```

2. ✅ ตรวจสอบเวลาในระบบ
   ```sql
   SELECT GETDATE(); -- SQL Server
   -- เปรียบเทียบกับ expire_at ใน Token table
   ```

3. ✅ แก้ไข token expiry duration (ถ้าจำเป็น)
   ```go
   // ใน email_service.go
   func GetVerificationTokenExpiry() time.Time {
       return time.Now().Add(24 * time.Hour) // ปรับเป็น 48 ชั่วโมง
   }
   ```

### Email Already Verified

**Symptoms:**
- ได้ข้อความ "email is already verified"

**This is expected behavior!**
- ตรวจสอบสถานะ:
  ```bash
  curl http://localhost:5000/auth/profile \
    -H "Authorization: Bearer YOUR_JWT_TOKEN"
  ```
- Response จะแสดง `"is_email_verified": true`

### Invalid Token

**Symptoms:**
- "invalid verification token" error

**Causes & Solutions:**
1. ✅ Token ถูก revoke แล้ว (ใช้ไปแล้ว)
   - ขอ token ใหม่ผ่าน resend API

2. ✅ Token ไม่มีในฐานข้อมูล
   ```sql
   SELECT * FROM Token 
   WHERE token = 'your-token-here' 
     AND token_type = 'email_verification';
   ```

3. ✅ Token URL ถูก encode ผิด
   - ตรวจสอบว่า URL ไม่มี spaces หรือ special characters ถูก encode

### Database Connection Issues

**Symptoms:**
- Errors เกี่ยวกับ database connection

**Solutions:**
1. ✅ ตรวจสอบ connection string
   ```env
   AZURESQL_SERVER=your-server.database.windows.net
   AZURESQL_DATABASE=Passion-tree-DB
   AZURESQL_USER=passion-tree-admin
   AZURESQL_PASSWORD=your-password
   ```

2. ✅ ตรวจสอบ Azure SQL Firewall rules
   - เพิ่ม IP address ของ server ใน Azure Portal

3. ✅ ทดสอบ connection
   ```bash
   # ใน terminal
   go run cmd/main.go
   # ดู console logs สำหรับ connection errors
   ```

## Future Enhancements
- [ ] Email template customization
- [ ] Multiple language support (Thai/English)
- [ ] Email verification reminder (after 24/48 hours)
- [ ] Optional email verification (configurable per environment)
- [ ] Admin panel to manually verify emails
- [ ] Rate limiting for resend verification (prevent abuse)
- [ ] Email verification analytics/metrics
- [ ] Custom email templates per user role

## Deployment Checklist

### Before Deployment
- [ ] ✅ Run database migrations (ALTER tables)
- [ ] ✅ Create indexes on Token table
- [ ] ✅ Set up SMTP credentials (production email service)
- [ ] ✅ Set APP_URL to production frontend URL
- [ ] ✅ Test email sending from production environment
- [ ] ✅ Verify firewall allows SMTP port (587/465)
- [ ] ✅ Set up SSL/TLS for email connection
- [ ] ✅ Configure proper FROM email address
- [ ] ✅ Test all API endpoints in staging

### Production Configuration
```env
# Production .env settings
SMTP_HOST=smtp.sendgrid.net          # หรือ email service ที่ใช้
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=SG.xxxxxxxxxxxxx
SMTP_FROM_EMAIL=noreply@passiontree.com
APP_URL=https://passiontree.com
```

### Monitoring & Maintenance

#### Scheduled Tasks
1. **Token Cleanup** (แนะนำรันทุกวัน)
   ```sql
   -- ลบ tokens ที่หมดอายุ
   DELETE FROM Token 
   WHERE expire_at < GETDATE() 
     AND token_type = 'email_verification';
   ```

2. **Email Bounce Tracking**
   - Monitor failed email deliveries
   - Update user records if email bounces

#### Metrics to Track
- จำนวน registrations ต่อวัน
- จำนวน email verifications ที่สำเร็จ
- Verification rate (verified/total registrations)
- Average time to verify email
- Failed email sending count
- Expired tokens count

#### Logs to Monitor
```go
// ตัวอย่าง log points
log.Printf("Verification email sent to: %s", email)
log.Printf("Email verified successfully: user_id=%s", userID)
log.Printf("Failed to send email: %v", err)
log.Printf("Token expired: token_id=%s", tokenID)
```

## Best Practices

1. **Email Content**
   - ใช้ภาษาที่เข้าใจง่าย
   - มีทั้งปุ่มและลิงก์สำรอง
   - แสดงเวลาหมดอายุชัดเจน
   - ใส่ contact information สำหรับ support

2. **Security**
   - ใช้ HTTPS สำหรับ verification links
   - อย่าแสดง token ใน logs
   - Implement rate limiting
   - Monitor suspicious activities

3. **User Experience**
   - แจ้งผู้ใช้ให้ check spam folder
   - มี UI แสดงสถานะ verification
   - Resend button ใช้งานง่าย
   - Redirect หลัง verify สำเร็จ

4. **Performance**
   - ใช้ async สำหรับส่งอีเมล
   - Cache email templates
   - Batch cleanup expired tokens
   - Monitor email sending queue

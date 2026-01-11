# JWT Authentication Implementation

## Overview
ระบบ authentication ของ Passion-Tree ใช้ JWT (JSON Web Token) สำหรับการยืนยันตัวตนผู้ใช้

## Features
- ✅ JWT Access Token (24 ชั่วโมง)
- ✅ Login ด้วย username หรือ email
- ✅ JWT Middleware สำหรับ protected routes
- ✅ User context จาก JWT claims

## API Endpoints

### Public Endpoints (ไม่ต้อง authentication)
- `POST /auth/register` - สมัครสมาชิก
- `POST /auth/login` - เข้าสู่ระบบ

### Protected Endpoints (ต้องส่ง JWT token)
- `GET /auth/profile` - ดูข้อมูลโปรไฟล์ตัวเอง
- `PUT /auth/profile` - แก้ไขโปรไฟล์ตัวเอง
- `PUT /auth/user` - แก้ไขข้อมูลผู้ใช้ตัวเอง
- `DELETE /auth/user` - ลบบัญชีตัวเอง

## ตัวอย่างการใช้งาน

### 1. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "test@example.com",
    "password": "testpass"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 2. ใช้ Token เข้าถึง Protected Endpoint
```bash
curl -X GET http://localhost:8080/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "user": {
      "user_id": "6D2C9B3F-8882-4746-8D33-33D96E1A82B3",
      "username": "testuser",
      "email": "test@example.com",
      ...
    },
    "profile": {
      ...
    }
  }
}
```

### 3. แก้ไขโปรไฟล์
```bash
curl -X PUT http://localhost:8080/auth/profile \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Hello World",
    "location": "Bangkok"
  }'
```

## JWT Token Structure

### Claims ที่เก็บใน Token:
```json
{
  "user_id": "uuid",
  "username": "string",
  "role": "user",
  "exp": 1234567890,
  "iat": 1234567890,
  "iss": "passion-tree"
}
```

## Environment Variables

เพิ่มใน `.env` file:
```env
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
```

⚠️ **สำคัญ**: ใน production ต้องตั้งค่า `JWT_SECRET` ให้เป็น random string ที่ยาวและซับซ้อน

## Security Notes

1. **Token Expiration**: Access token หมดอายุใน 24 ชั่วโมง
2. **HTTPS**: ใน production ต้องใช้ HTTPS เสมอ
3. **Secret Key**: ต้องเก็บ JWT_SECRET ให้ปลอดภัย
4. **Token Storage**: Client ควรเก็บ token ใน secure storage (ไม่ใช่ localStorage)

## การทดสอบ

### ทดสอบด้วย Postman:
1. Login เพื่อรับ token
2. Copy token
3. ตั้ง Authorization Type เป็น "Bearer Token"
4. Paste token แล้วเรียก protected endpoints

### ทดสอบด้วย curl:
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"test@example.com","password":"testpass"}' \
  | jq -r '.data.token')

# 2. ใช้ token
curl -X GET http://localhost:8080/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

## Files Structure

```
internal/
├── auth/
│   ├── jwt.go                 # JWT service (generate & validate)
│   ├── routes.go              # Route registration with middleware
│   ├── handler/
│   │   ├── handler.go
│   │   ├── user_hanler.go     # User handlers (login, register, etc.)
│   │   └── profile_handler.go # Profile handlers
│   ├── middleware/
│   │   └── jwt_middleware.go  # JWT authentication middleware
│   ├── service/
│   │   └── user_service.go    # Business logic with JWT
│   └── model/
│       └── user_model.go
```

## Next Steps (Optional)

- [ ] Implement Refresh Token
- [ ] Add token blacklist for logout
- [ ] Add rate limiting
- [ ] Add password reset functionality
- [ ] Add email verification
- [ ] Add OAuth2 (Google, Discord)

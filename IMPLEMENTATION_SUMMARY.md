# Search Integration - Implementation Summary

## Task: SCRUM-189-Search-Learning-Path

## Architecture Overview

**Flow:**
1. Frontend → Go Backend (`/api/v1/learningpaths/search`)
2. Go Backend → AI Service (`/search/`) - ส่ง query
3. AI Service → Qdrant - vector search
4. AI Service → Go Backend - **ส่งกลับเฉพาะ IDs + Scores**
5. Go Backend → Azure SQL - query รายละเอียดตาม IDs
6. Go Backend → Frontend - ส่งผลลัพธ์รวม (Scores + Details)

## ไฟล์ที่สร้างใหม่

### 1. AI Client Types (`internal/platform/aiclient/types.go`)
- แยก struct ออกจาก logic
- `SearchRequest`: รับ query, top_k, filters
- `SearchResult`: **เฉพาะ ID และ Score**
- `SearchResponse`: response จาก AI service

### 2. AI Client (`internal/platform/aiclient/client.go`)
- ใช้ **Fiber Client** แทน net/http
- method `Search()`: เรียก AI service
- method `Ping()`: ตรวจสอบ AI service
- Timeout และ error handling

### 3. Search Model (`internal/learning-path/model/search_model.go`)
- `SearchPathRequest`: รับจาก frontend
- `SearchPathResult`: **รวม Score + รายละเอียดเต็มจาก DB**
- `SearchPathResponse`: response ส่งกลับ frontend

### 4. Search Service (`internal/learning-path/service/search_service.go`)
- method `SearchLearningPaths()`:
  1. เรียก AI service เพื่อเอา IDs + Scores
  2. Loop query DB สำหรับแต่ละ ID
  3. รวม Score กับรายละเอียด
  4. ส่งกลับผลลัพธ์รวม
- Error handling สำหรับทั้ง AI และ DB

### 5. Search Handler (`internal/learning-path/handler/search_handler.go`)
- HTTP endpoint handler สำหรับ `/search`
- Validate request
- จัดการ response format

### 6. API Documentation (`SEARCH_API.md`)
- คู่มือการใช้งาน Search API
- อธิบาย architecture flow
- ตัวอย่าง request/response
- Configuration guide

## ไฟล์ที่แก้ไข

### 1. Config (`internal/config/config.go`)
- เพิ่ม field `AIServiceURL` ใน Config struct
- โหลด AI_SERVICE_URL จาก environment
- **Default value: `http://ai-service:8000`** (Docker Compose service name)

### 2. Service Interface (`internal/learning-path/service/service.go`)
- เพิ่ม `ServiceSearch` interface
- เพิ่ม `aiClient` field ใน serviceImpl
- อัปเดต `NewService()` รับ aiClient parameter

### 3. Handler (`internal/learning-path/handler/handler.go`)
- เพิ่ม `searchSvc` field
- Initialize ใน `NewHandler()`

### 4. Routes (`internal/learning-path/routes.go`)
- เพิ่ม parameter `aiClient` ใน `RegisterRoutes()`
- สร้าง aiClient instance และส่งต่อไปยัง service
- เพิ่ม route `POST /learningpaths/search`

### 5. Main Routes (`internal/routes/routes.go`)
- เพิ่ม import `aiclient`
- อัปเดต `Setup()` รับ aiClient parameter
- ส่ง aiClient ไปยัง `learningpath.RegisterRoutes()`

### 6. Main (`cmd/main.go`)
- เพิ่ม import `aiclient`
- สร้าง AI client instance จาก config
- ส่ง aiClient ไปยัง `routes.Setup()`

## API Endpoint ใหม่

```
POST /api/v1/learningpaths/search
```

### Request
```json
{
  "query": "Find beginner Go courses",
  "top_k": 7,
  "filters": {
    "category_id": 10
  }
}
```

### Response
```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "query": "Find beginner Go courses",
    "total": 5,
    "results": [
      {
        "path_id": "123",
        "score": 0.895,
        "title": "Go Fundamentals",
        "description": "Learn Go programming",
        "cover_img_url": "...",
        "metadata": {...}
      }
    ]
  }
}
```

## Environment Variables

เพิ่ม environment variable ใหม่:
```bash
# For Docker Compose (default)
AI_SERVICE_URL=http://ai-service:8000

# For local development
AI_SERVICE_URL=http://localhost:8000

# For production
AI_SERVICE_URL=https://your-ai-service.azurecontainerapps.io
```

## AI Service Changes (Python)

### Modified Files:
1. **`app/features/search/schemas.py`**
   - `SearchResult`: ลบ `payload` field ออก, เหลือเฉพาะ `id` และ `score`
   - อัปเดต example response

2. **`app/features/search/repository.py`**
   - `search()`: เปลี่ยน `with_payload=False`
   - Return เฉพาะ ID และ Score

### Response Format Change:
```python
# ก่อน (Old):
{
  "id": 123,
  "score": 0.895,
  "payload": {"title": "...", "category_id": 10}
}

# หลัง (New):
{
  "id": 123,
  "score": 0.895
}
```

## Testing

### Prerequisites
1. AI Service ต้องรันอยู่ที่ `AI_SERVICE_URL`
2. Qdrant มีข้อมูล Learning Paths

### How to Test
```bash
# Start AI Service
cd Passion-Tree-AI
python -m uvicorn app.main:app --reload

# Start Go Backend
cd Passion-Tree-Backend
go run cmd/main.go

# Test with curl
curl -X POST http://localhost:5000/api/v1/learningpaths/search \
  -H "Content-Type: application/json" \
  -d '{"query": "machine learning", "top_k": 5}'
```

## Integration Flow

1. Frontend → Go Backend (`/api/v1/learningpaths/search`)
2. Go Backend → AI Service (`/search/`) via **Fiber HTTP Client**
3. AI Service → Qdrant (vector search)
4. AI Service → Go Backend (**IDs + Scores only**)
5. Go Backend → Azure SQL Database (query details by IDs)
6. Go Backend (merge Scores + Details)
7. Go Backend → Frontend (final response with full details)

## Dependencies

### Go Backend:
- `github.com/gofiber/fiber/v2` - Web framework และ HTTP client

### AI Service:
- ไม่มี dependencies ใหม่ (ใช้ของเดิม)

## Key Improvements

1. **Separation of Concerns**: 
   - AI Service: จัดการเฉพาณ vector search
   - Go Backend: จัดการข้อมูลจริงจาก database

2. **Fiber HTTP Client**: 
   - ใช้ Fiber client แทน net/http
   - Consistency กับ web framework

3. **Type Organization**: 
   - แยก struct ออกเป็นไฟล์ `types.go`
   - Code organization ดีขึ้น

4. **Docker Ready**: 
   - ใช้ service name `ai-service` แทน `localhost`
   - พร้อมสำหรับ Docker Compose และ production

5. **Efficient Data Transfer**: 
   - AI ส่งเฉพาะ ID + Score (ข้อมูลน้อย)
   - Fetch รายละเอียดเต็มจาก DB เท่าที่ต้องการ

## Next Steps

1. ทดสอบ integration กับ AI service
2. ทดสอบกับ frontend
3. เพิ่ม error handling เพิ่มเติม
4. พิจารณาเพิ่ม caching (Redis)
5. เพิ่ม monitoring และ logging
6. เพิ่ม unit tests

## Notes

- **Fiber HTTP Client**: ใช้ `fiber.Client()` แทน `net/http.Client`
- **Default URL**: `http://ai-service:8000` (Docker service name)
- **Default top_k**: 7
- **Default resource_type**: "learning_paths"
- **AI Response**: เฉพาะ ID และ Score (ไม่มี payload)
- **Full Details**: Query จาก Azure SQL Database หลังจากได้ IDs
- **Error Handling**: ถ้า path ไม่มีใน DB จะ skip ไป

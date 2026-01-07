# Search API Integration

## Overview
ระบบ Search ที่เชื่อมต่อระหว่าง Go Backend และ AI Service (Python/FastAPI) โดย:
1. AI Service ทำ Semantic Search ใน Qdrant และส่งกลับเฉพาะ **ID และ Score**
2. Go Backend นำ ID ไป query รายละเอียดเต็มจาก **Azure SQL Database**
3. รวม Score กับรายละเอียดแล้วส่งกลับให้ Frontend

## Architecture

```
Frontend (Flutter/Dart)
    ↓
Go Backend (Fiber)
    ↓ (1) Search Request
AI Service (FastAPI + Qdrant)
    ↓ (2) Return: IDs + Scores
Go Backend
    ↓ (3) Query details by IDs
Azure SQL Database
    ↓ (4) Return: Full Path Details
Go Backend
    ↓ (5) Merge Score + Details
Frontend (Final Response)
```

## API Endpoint

### Search Learning Paths
**POST** `/api/v1/learningpaths/search`

#### Request Body
```json
{
  "query": "Find beginner Go courses",
  "top_k": 7,
  "filters": {
    "category_id": 10,
    "status": "active"
  }
}
```

#### Request Parameters
- `query` (string, required): ข้อความที่ต้องการค้นหา (Semantic Search)
- `top_k` (int, optional): จำนวนผลลัพธ์ที่ต้องการ (default: 7, max: 20)
- `filters` (object, optional): เงื่อนไขการกรองเพิ่มเติม (เช่น category_id, status)

#### Response
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
        "description": "Learn the basics of Go programming from scratch",
        "cover_img_url": "https://example.com/image.jpg",
        "objective": "Master Go basics",
        "avg_rating": 4.5,
        "status": "active",
        "creator_id": "user_001",
        "created_at": "2025-01-01T10:00:00Z",
        "updated_at": "2025-01-05T15:30:00Z"
      }
    ]
  }
}
```

#### Response Fields
- `query`: คำค้นหาที่ใช้
- `total`: จำนวนผลลัพธ์ทั้งหมด
- `results`: รายการผลลัพธ์
  - `path_id`: ID ของ Learning Path
  - `score`: คะแนนความเกี่ยวข้องจาก AI (0.0 - 1.0)
  - `title`: ชื่อ Learning Path (จาก DB)
  - `description`: คำอธิบาย (จาก DB)
  - `cover_img_url`: URL รูปภาพปก (จาก DB)
  - `objective`: วัตถุประสงค์ (จาก DB)
  - `avg_rating`: คะแนนเฉลี่ย (จาก DB)
  - `status`: สถานะ (จาก DB)
  - `creator_id`: ผู้สร้าง (จาก DB)
  - `created_at`: วันที่สร้าง (จาก DB)
  - `updated_at`: วันที่อัปเดต (จาก DB)

## Configuration

### Environment Variables
```bash
# AI Service URL - ใช้ service name สำหรับ Docker Compose
AI_SERVICE_URL=http://ai-service:8000

# For local development
AI_SERVICE_URL=http://localhost:8000

# For production
AI_SERVICE_URL=https://your-ai-service.azurecontainerapps.io
```

## File Structure

### Backend (Go)
```
internal/
  ├── platform/
  │   └── aiclient/
  │       ├── types.go            # Request/Response structs
  │       └── client.go           # Fiber HTTP client
  ├── learning-path/
  │   ├── model/
  │   │   └── search_model.go     # Search models
  │   ├── service/
  │   │   └── search_service.go   # Business logic (AI + DB)
  │   └── handler/
  │       └── search_handler.go   # HTTP handler
  └── config/
      └── config.go               # Configuration
```

### AI Service (Python)
```
app/
  ├── api/endpoints/
  │   └── search.py              # FastAPI endpoints
  └── features/search/
      ├── schemas.py             # Pydantic models (ID + Score only)
      ├── service.py             # Search logic
      └── repository.py          # Qdrant integration
```

## How It Works

1. **Client Request**: Frontend ส่ง search request มาที่ Go Backend
2. **AI Search**: Go Backend ส่ง query ไปยัง AI Service ผ่าน Fiber HTTP client
3. **Vector Search**: AI Service:
   - แปลง query เป็น embedding vector
   - ค้นหาใน Qdrant vector database
   - คำนวณ similarity score
   - **ส่งกลับเฉพาะ ID และ Score**
4. **Database Query**: Go Backend:
   - รับ IDs และ Scores จาก AI
   - Query รายละเอียดเต็มจาก Azure SQL Database สำหรับแต่ละ ID
   - รวม Score จาก AI กับรายละเอียดจาก DB
5. **Client Response**: ส่งผลลัพธ์รวมกลับไปยัง Frontend

## Key Features

- **Separation of Concerns**: AI Service รับผิดชอบเฉพาะ semantic search, DB รับผิดชอบข้อมูลจริง
- **Fiber HTTP Client**: ใช้ Fiber client แทน net/http เพื่อความสอดคล้องกับ framework
- **Type Safety**: แยก struct ออกเป็นไฟล์ types.go
- **Docker Ready**: ใช้ service name `ai-service` สำหรับ Docker Compose
- **Flexible Configuration**: รองรับทั้ง local development และ production deployment

## Example Usage

### Using cURL
```bash
curl -X POST http://localhost:5000/api/v1/learningpaths/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning for beginners",
    "top_k": 5
  }'
```

### Using Postman
1. Method: POST
2. URL: `http://localhost:5000/api/v1/learningpaths/search`
3. Headers: `Content-Type: application/json`
4. Body (raw JSON):
```json
{
  "query": "machine learning for beginners",
  "top_k": 5,
  "filters": {
    "status": "active"
  }
}
```

## Error Handling

### Common Errors

#### AI Service Unavailable
```json
{
  "success": false,
  "error": "failed to search via AI service: failed to send request: connection refused"
}
```
**Solution**: ตรวจสอบว่า AI Service กำลังรันอยู่และ AI_SERVICE_URL ถูกต้อง

#### Empty Query
```json
{
  "success": false,
  "error": "search query is required"
}
```
**Solution**: ระบุ query ใน request body

#### Invalid Request Body
```json
{
  "success": false,
  "error": "invalid request body"
}
```
**Solution**: ตรวจสอบ JSON format

## Testing

### Prerequisites
1. AI Service ต้อง running และ Qdrant มีข้อมูล Learning Paths
2. ตั้งค่า AI_SERVICE_URL ใน .env

### Test Steps
1. Start AI Service: `cd Passion-Tree-AI && python -m uvicorn app.main:app --reload`
2. Start Go Backend: `cd Passion-Tree-Backend && go run cmd/main.go`
3. Test search endpoint with Postman or cURL

## Performance Considerations

- **Timeout**: HTTP client timeout = 30 seconds
- **Caching**: พิจารณาเพิ่ม Redis cache สำหรับ popular queries
- **Rate Limiting**: พิจารณาเพิ่ม rate limiting ถ้า AI Service มี quota
- **Connection Pooling**: HTTP client ใช้ connection pooling โดย default

## Future Improvements

1. **Caching**: เพิ่ม Redis cache สำหรับผลลัพธ์การค้นหา
2. **Retry Logic**: เพิ่ม retry mechanism เมื่อ AI service ไม่ตอบสนอง
3. **Circuit Breaker**: ป้องกัน cascading failure
4. **Metrics**: เพิ่ม logging และ monitoring
5. **Search History**: บันทึก search queries เพื่อวิเคราะห์

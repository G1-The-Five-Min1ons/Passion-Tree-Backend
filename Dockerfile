# 1) Build stage
FROM golang:1.24-alpine AS builder

# ติดตั้ง dependencies ที่จำเป็นสำหรับการ Build
RUN apk add --no-cache git build-base

WORKDIR /src

# คัดลอกเฉพาะไฟล์ mod ก่อนเพื่อใช้ประโยชน์จาก Docker Cache
COPY go.mod go.sum ./
RUN go mod download

# คัดลอกโค้ดทั้งหมด
COPY . .

# ตั้งค่า Environment และ Build Binary
# ปรับเปลี่ยน path ปลายทางให้ชัดเจน
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -ldflags "-s -w" -o /out/app ./cmd

# 2) Runtime stage
FROM alpine:3.20

# ติดตั้งใบรับรองความปลอดภัยและเครื่องมือพื้นฐาน
RUN apk add --no-cache ca-certificates curl tzdata && update-ca-certificates

WORKDIR /app

# คัดลอก Binary มาจากขั้นตอน builder
COPY --from=builder /out/app ./app

# สร้าง User ใหม่ (ไม่ใช้ Root) เพื่อความปลอดภัยสูงสุด
RUN adduser -D -H -u 10001 appuser
USER 10001

# ตั้งค่าตัวแปรสภาพแวดล้อมพื้นฐาน
ENV PORT=5000
EXPOSE 5000

# รันแอปพลิเคชัน
ENTRYPOINT ["/app/app"]
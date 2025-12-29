package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// 1. สร้าง Interface เพื่อให้ Mock ได้ง่ายเวลาเขียน Test
type Database interface {
	GetDB() *sql.DB
	CheckConnection() error
	Close() error
}

// 2. ใช้ Struct เก็บสถานะ แทนการใช้ Global Variable
type sqlDatabase struct {
	db *sql.DB
}

// NewDatabase ทำหน้าที่เป็น Constructor รับ Connection String เข้ามา
// วิธีนี้ทำให้เราส่ง Connection String จำลอง (Mock) เข้ามาเทสได้ง่าย
func NewDatabase(connString string) (Database, error) {
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	// ตั้งค่า Pool เพื่อประสิทธิภาพ (Best Practice)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &sqlDatabase{db: db}, nil
}

// NewDatabaseWithRetry เชื่อมต่อ Database พร้อม Retry Logic สำหรับ Cloud DB
// maxRetries: จำนวนครั้งที่จะลองใหม่ (แนะนำ 5-10 ครั้ง)
// initialDelay: ระยะเวลารอครั้งแรก (แนะนำ 2-5 วินาที)
func NewDatabaseWithRetry(connString string, maxRetries int, initialDelay time.Duration) (Database, error) {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", attempt, maxRetries)

		db, err := NewDatabase(connString)
		if err == nil {
			log.Printf("Database connected successfully on attempt %d", attempt)
			return db, nil
		}

		lastErr = err
		log.Printf("Connection attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			log.Printf("Waiting %v before retry...", delay)
			time.Sleep(delay)
			// Exponential backoff: เพิ่มระยะเวลารอเป็น 2 เท่า (สูงสุด 30 วินาที)
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, lastErr)
}

func (s *sqlDatabase) GetDB() *sql.DB {
	return s.db
}

func (s *sqlDatabase) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}

func (s *sqlDatabase) Close() error {
	return s.db.Close()
}

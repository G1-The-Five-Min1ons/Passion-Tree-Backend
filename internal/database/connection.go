package database

import (
	"context"
	"database/sql"
	"fmt"
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
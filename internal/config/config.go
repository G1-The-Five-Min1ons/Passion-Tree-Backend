package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnString string
}

func LoadDBConfig() (*Config, error) {
	// โหลด .env เฉพาะตอนรันแอปจริง
	_ = godotenv.Load()

	server := os.Getenv("AZURESQL_SERVER")
	user := os.Getenv("AZURESQL_USER")
	password := os.Getenv("AZURESQL_PASSWORD")
	port := os.Getenv("AZURESQL_PORT")
	database := os.Getenv("AZURESQL_DATABASE")

	if server == "" || user == "" || password == "" || database == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	if port == "" {
		port = "1433" // ค่า Default ของ Azure SQL
	}

	// สร้าง Connection String เตรียมไว้
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=true",
		server, user, password, port, database)

	return &Config{
		DBConnString: connString,
	}, nil
}
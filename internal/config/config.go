package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnString            string
	AIServiceURL            string
	AzureStorageConnString  string
	ContainerLearningPath   string
	ContainerProfile        string
}

func LoadDBConfig() (*Config, error) {
	// โหลด .env เฉพาะตอนรันแอปจริง
	_ = godotenv.Load()

	server := os.Getenv("AZURESQL_SERVER")
	user := os.Getenv("AZURESQL_USER")
	password := os.Getenv("AZURESQL_PASSWORD")
	port := os.Getenv("AZURESQL_PORT")
	database := os.Getenv("AZURESQL_DATABASE")
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	
	// Azure Storage
	storageConnString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	containerLearningPath := os.Getenv("CONTAINER_LEARNING_PATH")
	containerProfile := os.Getenv("CONTAINER_PROFILE")

	if server == "" || user == "" || password == "" || database == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	if port == "" {
		port = "1433" // ค่า Default ของ Azure SQL
	}
	
	if containerLearningPath == "" {
		containerLearningPath = "learning-path-cover-imgs"
	}

	if containerProfile == "" {
		containerProfile = "profile-imgs"
	}

	if aiServiceURL == "" {
		aiServiceURL = "http://ai-service:8000" // Default for Docker Compose
	}

	// สร้าง Connection String เตรียมไว้
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=true",
		server, user, password, port, database)

	return &Config{
		DBConnString:           connString,
		AIServiceURL:           aiServiceURL,
		AzureStorageConnString: storageConnString,
		ContainerLearningPath:  containerLearningPath,
		ContainerProfile:       containerProfile,
	}, nil
}

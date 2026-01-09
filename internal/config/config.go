package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Default configuration values
const (
	DefaultAzureSQLPort      = "1433"
	DefaultAIServiceURL      = "http://ai-service:8000"
	DefaultContainerLearning = "learning-path-cover-imgs"
	DefaultContainerProfile  = "profile-imgs"
)

// Environment variable keys
const (
	EnvAzureSQLServer         = "AZURESQL_SERVER"
	EnvAzureSQLUser           = "AZURESQL_USER"
	EnvAzureSQLPassword       = "AZURESQL_PASSWORD"
	EnvAzureSQLPort           = "AZURESQL_PORT"
	EnvAzureSQLDatabase       = "AZURESQL_DATABASE"
	EnvAIServiceURL           = "AI_SERVICE_URL"
	EnvAzureStorageConnString = "AZURE_STORAGE_CONNECTION_STRING"
	EnvContainerLearningPath  = "CONTAINER_LEARNING_PATH"
	EnvContainerProfile       = "CONTAINER_PROFILE"
)

type Config struct {
	DBConnString           string
	AIServiceURL           string
	AzureStorageConnString string
	ContainerLearningPath  string
	ContainerProfile       string
}

// LoadDBConfig loads configuration from environment variables
func LoadDBConfig() (*Config, error) {
	// Load .env file (ignores error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		AIServiceURL:           getEnvOrDefault(EnvAIServiceURL, DefaultAIServiceURL),
		AzureStorageConnString: os.Getenv(EnvAzureStorageConnString),
		ContainerLearningPath:  getEnvOrDefault(EnvContainerLearningPath, DefaultContainerLearning),
		ContainerProfile:       getEnvOrDefault(EnvContainerProfile, DefaultContainerProfile),
	}

	// Build database connection string
	connString, err := buildDBConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to build database connection string: %w", err)
	}
	config.DBConnString = connString

	return config, nil
}

// buildDBConnectionString constructs the database connection string from environment variables
func buildDBConnectionString() (string, error) {
	server := os.Getenv(EnvAzureSQLServer)
	user := os.Getenv(EnvAzureSQLUser)
	password := os.Getenv(EnvAzureSQLPassword)
	database := os.Getenv(EnvAzureSQLDatabase)
	port := getEnvOrDefault(EnvAzureSQLPort, DefaultAzureSQLPort)

	if server == "" || user == "" || password == "" || database == "" {
		return "", fmt.Errorf("missing required database environment variables")
	}

	return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=true",
		server, user, password, port, database), nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

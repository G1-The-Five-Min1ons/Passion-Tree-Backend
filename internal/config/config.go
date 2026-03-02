package config

import (
	"fmt"
	"os"
	"log/slog"

	"github.com/joho/godotenv"
)

// Default configuration values
const (
	DefaultAzureSQLPort      = "1433"
	DefaultAIServiceURL      = "http://ai-service:8000"
	DefaultContainerLearning  = "learning-path-cover-imgs"
	DefaultContainerProfile   = "profile-imgs"
	DefaultContainerReflect   = "reflect"
	DefaultContainerMaterials = "materials-nodes"
	DefaultJWTAccessTTL      = "1"   // 1 hour 
    DefaultJWTRefreshTTL     = "168" // 7 days
    DefaultJWTRefreshAbsolute = "720" // 30 days
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
	EnvContainerReflect       = "CONTAINER_REFLECT"
	EnvContainerMaterials     = "CONTAINER_MATERIALS_NODES"
	EnvSMTPHost               = "SMTP_HOST"
	EnvSMTPPort               = "SMTP_PORT"
	EnvSMTPUsername           = "SMTP_USERNAME"
	EnvSMTPPassword           = "SMTP_PASSWORD"
	EnvSMTPFromEmail          = "SMTP_FROM_EMAIL"
	EnvMailerSendAPIKey       = "MAILERSEND_API_KEY"
	EnvAppURL                 = "APP_URL"

	EnvGmailEmail         = "GMAIL_EMAIL"
    EnvGmailAppPassword   = "GMAIL_APP_PASSWORD"

	// OAuth Environment variables

	// JWT Environment variables

	EnvJWTSecret           = "JWT_SECRET"
	EnvJWTAccessTTL        = "JWT_ACCESS_TTL"
	EnvJWTRefreshTTL       = "JWT_REFRESH_TTL"
	EnvJWTRefreshAbsolute  = "JWT_REFRESH_ABSOLUTE"
	EnvGoogleClientID      = "GOOGLE_CLIENT_ID"
	EnvGoogleClientSecret  = "GOOGLE_CLIENT_SECRET"
	EnvGoogleRedirectURL   = "GOOGLE_REDIRECT_URL"
	EnvDiscordClientID     = "DISCORD_CLIENT_ID"
	EnvDiscordClientSecret = "DISCORD_CLIENT_SECRET"
	EnvDiscordRedirectURL  = "DISCORD_REDIRECT_URL"
)

type Config struct {
	DBConnString           string
	AIServiceURL           string
	AzureStorageConnString string
	ContainerLearningPath    string
	ContainerProfile         string
	ContainerReflect         string
	ContainerMaterialsNodes  string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFromEmail          string
	MailerSendAPIKey       string
	AppURL                 string

	GmailEmail       		string
    GmailAppPassword 		string

	// OAuth settings
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string

	// JWT settings
	JWTSecret          string
	JWTAccessTTL       string // in hours
	JWTRefreshTTL      string // in hours (sliding window)
	JWTRefreshAbsolute string // in hours (absolute maximum)
}

// LoadDBConfig loads configuration from environment variables
func LoadDBConfig() (*Config, error) {
	// Load .env file (ignores error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		AIServiceURL:           getEnvOrDefault(EnvAIServiceURL, DefaultAIServiceURL),
		AzureStorageConnString: os.Getenv(EnvAzureStorageConnString),
		ContainerLearningPath:  getEnvOrDefault(EnvContainerLearningPath, DefaultContainerLearning),

		// OAuth settings
		GoogleClientID:      os.Getenv(EnvGoogleClientID),
		GoogleClientSecret:  os.Getenv(EnvGoogleClientSecret),
		GoogleRedirectURL:   getEnvOrDefault(EnvGoogleRedirectURL, "http://localhost:5000/auth/google/callback"),
		DiscordClientID:     os.Getenv(EnvDiscordClientID),
		DiscordClientSecret: os.Getenv(EnvDiscordClientSecret),
		DiscordRedirectURL:  getEnvOrDefault(EnvDiscordRedirectURL, "http://localhost:5000/auth/discord/callback"),

		// JWT settings
		JWTSecret:     os.Getenv(EnvJWTSecret),
		JWTAccessTTL:  getEnvOrDefault(EnvJWTAccessTTL, DefaultJWTAccessTTL),
		JWTRefreshTTL: getEnvOrDefault(EnvJWTRefreshTTL, DefaultJWTRefreshTTL),
		JWTRefreshAbsolute: getEnvOrDefault(EnvJWTRefreshAbsolute, DefaultJWTRefreshAbsolute),

		ContainerProfile:        getEnvOrDefault(EnvContainerProfile, DefaultContainerProfile),
		ContainerReflect:        getEnvOrDefault(EnvContainerReflect, DefaultContainerReflect),
		ContainerMaterialsNodes: getEnvOrDefault(EnvContainerMaterials, DefaultContainerMaterials),
		SMTPHost:         os.Getenv(EnvSMTPHost),
		SMTPPort:         getEnvOrDefault(EnvSMTPPort, "587"),
		SMTPUsername:     os.Getenv(EnvSMTPUsername),
		SMTPPassword:     os.Getenv(EnvSMTPPassword),
		SMTPFromEmail:    os.Getenv(EnvSMTPFromEmail),
		MailerSendAPIKey: os.Getenv(EnvMailerSendAPIKey),
		AppURL:           getEnvOrDefault(EnvAppURL, "http://localhost:5000"),

		GmailEmail:       os.Getenv(EnvGmailEmail),
        GmailAppPassword: os.Getenv(EnvGmailAppPassword),
	}

	// Build database connection string
	connString, err := buildDBConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to build database connection string: %w", err)
	}
	config.DBConnString = connString

	// Validate required JWT secret
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required in environment variables")
	}

	if config.GoogleClientID == "" || config.DiscordClientID == "" {
        slog.Warn("Missing Google or Discord client ID in configuration")
    }

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

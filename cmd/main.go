package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/robfig/cron/v3"

	authrepo "passiontree/internal/auth/repository"
	authservice "passiontree/internal/auth/service"
	"passiontree/internal/config"
	"passiontree/internal/connection"
	"passiontree/internal/pkg/logger"
	"passiontree/internal/pkg/migration"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/routes"
	"passiontree/internal/worker"
	workerRepo "passiontree/internal/worker/repository"

	missionRepo "passiontree/internal/mission/repository"
	missionService "passiontree/internal/mission/service"

	pathrepo "passiontree/internal/learning-path/repository"
	recrepo "passiontree/internal/recommendation/repository"
	recservice "passiontree/internal/recommendation/service"
)

const (
	DefaultPort     = "5000"
	DBRetryAttempts = 10
	DBRetryDelay    = 3 * time.Second
	AppName         = "Passion Tree Backend v1.0"
)

func main() {
	// Setup
	isDev := os.Getenv("APP_ENV") != "production"
	myLogger := logger.SetupLogger(isDev)

	cfg, err := config.LoadDBConfig()
	if err != nil {
		myLogger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize database connection
	db, err := initializeDatabase(cfg.DBConnString, myLogger)
	if err != nil {
		myLogger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run database migrations
	if err := runMigrations(db, myLogger); err != nil {
		myLogger.Error("Failed to run migrations", "error", err)
		myLogger.Warn("Continuing despite migration error - tables may already exist")
	}

	aiClient := initializeAIClient(cfg.AIServiceURL, myLogger) // Initialize AI client
	storageClient := initializeStorageClient(cfg, myLogger)    // Initialize Azure Storage client

	// Setup Fiber with custom Logger
	app := createFiberApp(myLogger)
	emailService := authservice.NewEmailService(cfg, myLogger)
	cronJob, notificationWorker := initializeBackgroundJobs(db, storageClient, emailService, aiClient, myLogger)
	routes.Setup(app, db, aiClient, storageClient, notificationWorker, myLogger)
	defer cronJob.Stop()

	port := getPort()
	myLogger.Info("starting server", "port", port, "app_name", AppName)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			myLogger.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	myLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		myLogger.Error("Fiber shutdown error", "error", err)
	}

	cronJob.Stop()
	myLogger.Info("Server exited gracefully")
}

// initializeDatabase connects to the database with retry logic
func initializeDatabase(connString string, logger *slog.Logger) (connection.Database, error) {
	db, err := connection.NewDatabaseWithRetry(connString, DBRetryAttempts, DBRetryDelay)
	if err != nil {
		return nil, err
	}
	logger.Info("database connected successfully")
	return db, nil
}

// runMigrations runs database migrations for social auth
func runMigrations(db connection.Database, logger *slog.Logger) error {
	migrator := migration.NewMigrator(db.GetDB(), logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return migrator.RunAllMigrations(ctx)
}

// initializeAIClient creates and configures the AI service client
func initializeAIClient(serviceURL string, logger *slog.Logger) *aiclient.AIClient {
	if serviceURL == "" {
		logger.Error("AI_SERVICE_URL is not set or empty. Please check your environment variables or configuration.")
		os.Exit(1)
	}
	client := aiclient.NewAIClient(serviceURL)
	logger.Info("AI Service configured", "service_url", serviceURL)
	return client
}

// initializeStorageClient creates Azure Storage client if configured
func initializeStorageClient(cfg *config.Config, logger *slog.Logger) *storage.BlobService {
	if cfg.AzureStorageConnString == "" {
		logger.Info("Azure Storage not configured, skipping initialization")
		return nil
	}

	storageClient, err := connection.InitBlobStorage(cfg)
	if err != nil {
		logger.Warn("Failed to initialize Azure Storage", "error", err)
		return nil
	}

	logger.Info("Azure Storage initialized successfully")
	return storageClient
}

// createFiberApp creates and configures the Fiber application with middleware
func createFiberApp(logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:     AppName,
		ProxyHeader: "X-Forwarded-For",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.ErrorContext(c.UserContext(), "unhandled_framework_error",
				"err", err.Error(),
				"method", c.Method(),
				"path", c.Path(),
				"ip", c.IP(),
			)

			return fiber.DefaultErrorHandler(c, err)
		},
	})

	app.Use(flogger.New())
	app.Use(fiberRecover.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == "https://passion-tree.org" || origin == "http://localhost:3000" {
				return true
			}

			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
				return true
			}

			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	logger.Info("fiber_application_initialized", "app_name", AppName)

	return app
}

// getPort returns the server port from environment or default
func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return DefaultPort
}

func initializeBackgroundJobs(db connection.Database, storage *storage.BlobService, emailService authservice.EmailService, aiClient *aiclient.AIClient, logger *slog.Logger) (*cron.Cron, *worker.EmailNotificationWorker) {
	cleanupWorker := worker.NewCleanupWorker(db, storage)
	notificationProvider := workerRepo.NewSQLEmailNotificationDataProvider(db)
	authRepository := authrepo.NewRepository(db)
	userService := authservice.NewUserService(authRepository, nil, nil, nil, logger)
	notificationWorker := worker.NewEmailNotificationWorker(notificationProvider, emailService, userService, logger)
	mRepo := missionRepo.NewRepository(db)
	mSvc := missionService.NewService(mRepo, authRepository, logger)
	pathRepository := pathrepo.NewRepository(db)
	recRepository := recrepo.NewRepository(db)
	recSvc := recservice.NewService(recRepository, pathRepository, aiClient, logger)

	c := cron.New()

	// Run every midnight
	_, err := c.AddFunc("0 0 * * *", withSafeCron(
		"StorageCleanup", 
		30*time.Minute, // ให้เวลา Cleanup สูงสุด 30 นาที
		logger, 
		func(ctx context.Context) error {
			cleanupWorker.RunCleanup()
			return nil
		},
	))
	if err != nil {
		logger.Error("error initializing background jobs", "error", err)
		return c, notificationWorker
	}

	_, err = c.AddFunc("0 2 * * *", withSafeCron(
		"DailyRecommendationBatch",
		30*time.Minute,
		logger,
		func(ctx context.Context) error {
			return recSvc.RunDailyRecommendationBatch(ctx)
		},
	))
	if err != nil {
		logger.Error("error initializing daily recommendation batch job", "error", err)
	}

	_, err = c.AddFunc("0 0 * * 1", withSafeCron(
		"AutoAssignWeeklyMissions",
		10*time.Minute,
		logger,
		func(ctx context.Context) error {
			return mSvc.AutoAssignWeeklyMissions(ctx)
		},
	))
	if err != nil {
		logger.Error("error initializing weekly mission assignment job", "error", err)
	}

	_, err = c.AddFunc("0 0 * * *", withSafeCron(
		"CleanupExpiredMissions", 
		5*time.Minute, 
		logger, 
		func(ctx context.Context) error {
			return mSvc.CleanupExpiredMissions(ctx)
		},
	))
	if err != nil {
		logger.Error("error initializing cleanup expired missions job", "error", err)
	}

	// Run daily notifications at 08:00
	_, err = c.AddFunc("0 8 * * *", withSafeCron(
		"DailyNotifications", 
		15*time.Minute, 
		logger, 
		func(ctx context.Context) error {
			notificationWorker.RunDailyNotifications()
			return nil
		},
	))
	if err != nil {
		logger.Error("error initializing daily notification job", "error", err)
		return c, notificationWorker
	}

	// Run weekly notifications every Monday at 09:00
	_, err = c.AddFunc("0 9 * * 1", withSafeCron(
		"WeeklyNotifications", 
		15*time.Minute, 
		logger, 
		func(ctx context.Context) error {
			notificationWorker.RunWeeklyNotifications()
			return nil
		},
	))
	if err != nil {
		logger.Error("error initializing weekly notification job", "error", err)
		return c, notificationWorker
	}

	c.Start()
	logger.Info("background jobs started")
	return c, notificationWorker
}

// withSafeCron เป็น Wrapper ช่วยจัดการ Timeout และ Panic ให้กับ Cron Job
func withSafeCron(jobName string, timeout time.Duration, logger *slog.Logger, task func(ctx context.Context) error) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("CronJob Panic Recovered (Prevented Server Crash)", "job_name", jobName, "panic_info", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		logger.Info("CronJob Started", "job_name", jobName)

		if err := task(ctx); err != nil {
			logger.Error("CronJob Failed", "job_name", jobName, "error", err)
		} else {
			logger.Info("CronJob Completed Successfully", "job_name", jobName)
		}
	}
}

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	
	"passiontree/API/internal/recommendation"
	"passiontree/API/pkg/database"
)

func main() {
	cfg := struct {
		User string
		Pass string
		Host string
		Name string
	}{
		User: "admin_user",
		Pass: "SecretPass123!",
		Host: "myserver.database.windows.net",
		Name: "my_shop_db",
	}

	db, err := database.NewAzureSQL(cfg.User, cfg.Pass, cfg.Host, cfg.Name)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer db.Close()
	log.Println("Connected to Azure SQL successfully!")

	recRepo := recommendation.NewRepository(db)
	recSvc := recommendation.NewService(recRepo)
	recHandler := recommendation.NewHandler(recSvc)

	r := gin.Default()

	apiV1 := r.Group("/api/v1")
	{
		recHandler.RegisterRoutes(apiV1)
	}

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

const DefaultPort = "8000"

func main() {

	app := fiber.New() 
	
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.Status(fiber.StatusOK).JSON(fiber.Map{
            "status": "up",
            "service": "Go Backend Orchestrator",
        })
    })
    
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	log.Printf("Starting Fiber server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
package main

import (
	"log"
	"skribbl-clone/config"
	"skribbl-clone/db"
	"skribbl-clone/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize Fiber
	app := fiber.New()

	// Load configuration
	config.LoadConfig()
	db.InitDB()

	// Initialize routes
	routes.SetupRoutes(app)

	// Start the server
	log.Println("Starting server on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

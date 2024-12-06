package routes

import (
	"skribbl-clone/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Game-related routes
	api := app.Group("/api")
	api.Get("/games", handlers.GetGames)
	api.Post("/games", handlers.CreateGame)
	api.Get("/games/:id", handlers.GetGameByID)
}

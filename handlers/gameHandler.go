package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func GetGames(c *fiber.Ctx) error {
	return nil
}

func CreateGame(c *fiber.Ctx) error {
	return nil

}

func GetGameByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// Fetch game by ID
	return c.JSON(fiber.Map{"status": "success", "message": "Game found", "id": id})
}

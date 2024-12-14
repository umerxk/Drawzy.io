package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	err2 := godotenv.Load()
	if err2 != nil {
		log.Println("No .env file found. Using defaults.")
	}
}

package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey    string
	APISecret string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, reading from environment variables")
	}

	key := os.Getenv("BINANCE_API_KEY")
	secret := os.Getenv("BINANCE_API_SECRET")

	if key == "" || secret == "" {
		return nil, fmt.Errorf("BINANCE_API_KEY and BINANCE_API_SECRET must be set in .env or environment")
	}

	return &Config{APIKey: key, APISecret: secret}, nil
}

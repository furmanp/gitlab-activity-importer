package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get the user home directory:", err)
	}
	return homeDir
}

func LoadEnv() error {

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath := wd + "/.env"

	if err := godotenv.Load(envPath); err != nil {

		return fmt.Errorf("failed to load .env file from %s: %w", envPath, err)
	}

	return nil
}

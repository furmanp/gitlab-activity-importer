package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

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
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("could not get caller info")
	}

	currentDir := filepath.Dir(filename)
	envPath := filepath.Join(currentDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		return fmt.Errorf("error loading .env file from %s: %w", envPath, err)
	}
	return nil
}

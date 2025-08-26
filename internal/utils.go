package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func CheckEnvVariables() error {
	requiredEnvVars := []string{
		"BASE_URL",
		"GITLAB_TOKEN",
		"GITLAB_USERNAME",
		"GH_USERNAME",
		"COMMITER_EMAIL",
		"ORIGIN_REPO_URL",
		"ORIGIN_TOKEN",
	}

	var missingVars []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}
	return nil
}

func LoadEnv() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath := filepath.Join(wd, ".env")

	if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error loading .env file from %s: %w", envPath, err)
	}

	return nil
}

func SetupEnv() error {
	if err := LoadEnv(); err != nil {
		log.Printf("Could not load .env file: %v.", err)
	}

	if err := CheckEnvVariables(); err != nil {
		return fmt.Errorf("environment variable check failed: %w", err)
	}

	return nil
}

func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get the user home directory:", err)
	}
	return homeDir
}

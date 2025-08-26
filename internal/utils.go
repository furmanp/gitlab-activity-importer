package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv tries to load .env from the project root.
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

// CheckEnvVariables ensures required env variables are present.
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

// SetupEnv loads .env if DEVELOPMENT, then checks variables.
func SetupEnv() error {
	if os.Getenv("ENV") == "DEVELOPMENT" {

		if err := LoadEnv(); err != nil {
			return fmt.Errorf("failed to load .env in DEVELOPMENT: %w", err)
		}
	}

	return CheckEnvVariables()
}

// GetHomeDirectory is unrelated but kept as-is.
func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get the user home directory:", err)
	}
	return homeDir
}

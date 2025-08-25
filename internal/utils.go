package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func CheckEnvVariables() error {
	if os.Getenv("ENV") == "DEVELOPMENT" {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("error loading .env file: %v", err)
		}
	}

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

func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get the user home directory:", err)
	}
	return homeDir
}

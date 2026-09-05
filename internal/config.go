package internal

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BaseURL        string
	GitlabUsername string
	GitlabToken    string
	GithubUsername string
	CommitterEmail string
	OriginRepoURL  string
	OriginToken    string
}

func Load() (Config, error) {
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
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return Config{
		BaseURL:        os.Getenv("BASE_URL"),
		GitlabUsername: os.Getenv("GITLAB_USERNAME"),
		GitlabToken:    os.Getenv("GITLAB_TOKEN"),
		GithubUsername: os.Getenv("GH_USERNAME"),
		CommitterEmail: os.Getenv("COMMITER_EMAIL"),
		OriginRepoURL:  os.Getenv("ORIGIN_REPO_URL"),
		OriginToken:    os.Getenv("ORIGIN_TOKEN"),
	}, nil
}

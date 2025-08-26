package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

func TestEnvironmentVariables(t *testing.T) {
	envVars := []string{
		"BASE_URL", "GITLAB_TOKEN", "GITLAB_USERNAME",
		"GH_USERNAME", "COMMITER_EMAIL", "ORIGIN_REPO_URL", "ORIGIN_TOKEN",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}

	defer func() {
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}()

	t.Run("All required environment variables present", func(t *testing.T) {
		os.Setenv("BASE_URL", "https://gitlab.com")
		os.Setenv("GITLAB_TOKEN", "token")
		os.Setenv("GITLAB_USERNAME", "gitlab_user")
		os.Setenv("GH_USERNAME", "github_user")
		os.Setenv("COMMITER_EMAIL", "user@example.com")
		os.Setenv("ORIGIN_REPO_URL", "https://github.com/user/repo.git")
		os.Setenv("ORIGIN_TOKEN", "token")

		err := internal.SetupEnv()

		if err != nil {
			t.Errorf("Expected no error when all variables are set, got: %v", err)
		}
	})

	t.Run("Missing GITLAB_USERNAME", func(t *testing.T) {
		// Clear all variables first
		for _, v := range envVars {
			os.Unsetenv(v)
		}

		os.Setenv("BASE_URL", "https://gitlab.com")
		os.Setenv("GITLAB_TOKEN", "token")
		os.Setenv("GH_USERNAME", "github_user")
		os.Setenv("COMMITER_EMAIL", "user@example.com")
		os.Setenv("ORIGIN_REPO_URL", "https://github.com/user/repo.git")
		os.Setenv("ORIGIN_TOKEN", "token")

		err := internal.SetupEnv()

		if err == nil {
			t.Error("Expected error when GITLAB_USERNAME is missing")
		}

		if !strings.Contains(err.Error(), "GITLAB_USERNAME") {
			t.Errorf("Expected error to mention GITLAB_USERNAME, got: %v", err)
		}
	})
}

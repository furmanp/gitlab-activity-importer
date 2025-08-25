package integration_test

import (
	"testing"
	"os"
	"strings"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

// TestEnvironmentVariableValidation tests that both usernames are required
func TestEnvironmentVariableValidation(t *testing.T) {
	// Clear env vars
	envVars := []string{"BASE_URL", "GITLAB_TOKEN", "GITLAB_USERNAME", "GITHUB_USERNAME", "COMMITER_EMAIL", "ORIGIN_REPO_URL", "ORIGIN_TOKEN"}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
	defer func() {
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}()

	// Test missing GITLAB_USERNAME
	os.Setenv("BASE_URL", "https://gitlab.com")
	os.Setenv("GITLAB_TOKEN", "token")
	os.Setenv("GITHUB_USERNAME", "GitHub User")
	os.Setenv("COMMITER_EMAIL", "user@example.com")
	os.Setenv("ORIGIN_REPO_URL", "https://github.com/user/repo.git")
	os.Setenv("ORIGIN_TOKEN", "token")
	
	err := internal.CheckEnvVariables()
	if err == nil {
		t.Error("Expected error when GITLAB_USERNAME is missing")
	}
	if !strings.Contains(err.Error(), "GITLAB_USERNAME") {
		t.Errorf("Expected error to mention GITLAB_USERNAME, got: %v", err)
	}
	
	// Test with all variables set
	os.Setenv("GITLAB_USERNAME", "gitlab_user")
	err = internal.CheckEnvVariables()
	if err != nil {
		t.Errorf("Expected no error when all variables are set, got: %v", err)
	}
	
	t.Log("✅ Environment variable validation test passed")
}

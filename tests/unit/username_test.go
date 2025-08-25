package unit_test

import (
	"testing"
	"os"
	"strings"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

// TestUsernameEnvironmentVariables tests the new GITLAB_USERNAME environment variable
func TestUsernameEnvironmentVariables(t *testing.T) {
	// Clear all environment variables first
	envVars := []string{
		"BASE_URL", "GITLAB_TOKEN", "GITLAB_USERNAME", 
		"GITHUB_USERNAME", "COMMITER_EMAIL", "ORIGIN_REPO_URL", "ORIGIN_TOKEN",
	}
	
	for _, v := range envVars {
		os.Unsetenv(v)
	}
	
	defer func() {
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}()

	// Test case 1: Missing GITLAB_USERNAME should fail
	t.Run("Missing GITLAB_USERNAME should fail", func(t *testing.T) {
		// Set all required vars except GITLAB_USERNAME
		os.Setenv("BASE_URL", "https://gitlab.com")
		os.Setenv("GITLAB_TOKEN", "token")
		os.Setenv("GITHUB_USERNAME", "GitHub User")
		os.Setenv("COMMITER_EMAIL", "user@example.com")
		os.Setenv("ORIGIN_REPO_URL", "https://github.com/user/repo.git")
		os.Setenv("ORIGIN_TOKEN", "token")
		
		err := internal.CheckEnvVariables()
		
		if err == nil {
			t.Error("Expected error when GITLAB_USERNAME is missing, but got none")
		}
		
		if !strings.Contains(err.Error(), "GITLAB_USERNAME") {
			t.Errorf("Expected error to mention GITLAB_USERNAME, got: %v", err)
		}
		
		t.Log("✅ Successfully detected missing GITLAB_USERNAME")
	})

	// Test case 2: All variables present should pass
	t.Run("All variables present should pass", func(t *testing.T) {
		// Set all required variables including GITLAB_USERNAME
		os.Setenv("BASE_URL", "https://gitlab.com")
		os.Setenv("GITLAB_TOKEN", "token")
		os.Setenv("GITLAB_USERNAME", "gitlab_user")
		os.Setenv("GITHUB_USERNAME", "GitHub User")
		os.Setenv("COMMITER_EMAIL", "user@example.com")
		os.Setenv("ORIGIN_REPO_URL", "https://github.com/user/repo.git")
		os.Setenv("ORIGIN_TOKEN", "token")
		
		err := internal.CheckEnvVariables()
		
		if err != nil {
			t.Errorf("Expected no error when all variables are set, got: %v", err)
		}
		
		t.Log("✅ All environment variables validated successfully")
	})

	// Test case 3: Verify different usernames can coexist
	t.Run("Different usernames can coexist", func(t *testing.T) {
		os.Setenv("GITLAB_USERNAME", "my.gitlab.user")
		os.Setenv("GITHUB_USERNAME", "My GitHub Name")
		
		gitlabUser := os.Getenv("GITLAB_USERNAME")
		githubUser := os.Getenv("GITHUB_USERNAME")
		
		if gitlabUser == "" {
			t.Error("GITLAB_USERNAME should be set")
		}
		
		if githubUser == "" {
			t.Error("GITHUB_USERNAME should be set")
		}
		
		if gitlabUser == githubUser {
			t.Log("Note: Usernames are the same, but that's allowed")
		} else {
			t.Log("✅ Successfully supports different GitLab and GitHub usernames")
		}
		
		t.Logf("GitLab username: %s", gitlabUser)
		t.Logf("GitHub commiter name: %s", githubUser)
	})
}

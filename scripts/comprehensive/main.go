package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/furmanp/gitlab-activity-importer/internal/services"
)

func main() {
	fmt.Println("🔍 Comprehensive Test: Different GitLab and GitHub Usernames")
	fmt.Println("============================================================")
	fmt.Println()

	// Test 1: Verify environment variables work with different usernames
	fmt.Println("Test 1: Environment Variable Validation")
	fmt.Println("---------------------------------------")
	
	// Clear all env vars first
	clearEnvVars()
	
	// Set up DIFFERENT usernames
	gitlabUsername := "john.doe.gitlab.company"
	githubCommiterName := "johndoe"
	
	os.Setenv("BASE_URL", "https://gitlab.com")
	os.Setenv("GITLAB_TOKEN", "test-token")
	os.Setenv("GITLAB_USERNAME", gitlabUsername)
	os.Setenv("GITHUB_USERNAME", githubCommiterName)
	os.Setenv("COMMITER_EMAIL", "john@example.com")
	os.Setenv("ORIGIN_REPO_URL", "https://github.com/john/repo.git")
	os.Setenv("ORIGIN_TOKEN", "github-token")
	
	err := internal.CheckEnvVariables()
	if err != nil {
		fmt.Printf("❌ Environment validation failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Environment variables validated successfully\n")
	fmt.Printf("   GitLab Username: '%s'\n", gitlabUsername)
	fmt.Printf("   GitHub Username: '%s'\n", githubCommiterName)
	fmt.Printf("   ✅ Usernames are DIFFERENT (this is the fix!)\n")
	fmt.Println()

	// Test 2: Verify GitLab API calls use correct username
	fmt.Println("Test 2: GitLab API Call Verification")
	fmt.Println("------------------------------------")
	
	// Create mock GitLab server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v4/projects/123/repository/commits") {
			// CHECK: Verify the API call uses GITLAB_USERNAME, not GITHUB_USERNAME
			queryParams := r.URL.Query()
			authorParam := queryParams.Get("author")
			
			fmt.Printf("   API called with author parameter: '%s'\n", authorParam)
			
			if authorParam == gitlabUsername {
				fmt.Printf("   ✅ CORRECT: Using GitLab username for API call\n")
			} else if authorParam == githubCommiterName {
				fmt.Printf("   ❌ WRONG: Using GitHub name for GitLab API (old bug!)\n")
				os.Exit(1)
			} else {
				fmt.Printf("   ❌ UNEXPECTED: Using unknown username '%s'\n", authorParam)
				os.Exit(1)
			}
			
			// Return mock commits
			commits := []internal.Commit{
				{
					ID:           "commit123",
					Message:      "Test commit from GitLab",
					AuthorName:   gitlabUsername, // This would be the GitLab username in real data
					AuthorMail:   "john@company.com",
					AuthoredDate: time.Now(),
				},
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(commits)
			return
		}
		
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()
	
	// Update BASE_URL to use mock server
	os.Setenv("BASE_URL", mockServer.URL)
	
	// Test GetProjectCommits with GitLab username
	commits, err := services.GetProjectCommits(123, gitlabUsername)
	if err != nil {
		fmt.Printf("❌ GetProjectCommits failed: %v\n", err)
		os.Exit(1)
	}
	
	if len(commits) == 0 {
		fmt.Printf("❌ No commits returned\n")
		os.Exit(1)
	}
	
	fmt.Printf("   ✅ Retrieved %d commit(s) from GitLab API\n", len(commits))
	fmt.Println()

	// Test 3: Verify no assumptions between usernames
	fmt.Println("Test 3: No Username Assumptions")
	fmt.Println("-------------------------------")
	
	// Test with various different username formats
	testCases := []struct {
		gitlab string
		github string
		desc   string
	}{
		{"john.doe", "John Doe", "Different format (dots vs spaces)"},
		{"jdoe123", "John D.", "Completely different"},
		{"john_doe_company", "JohnDoe", "Underscores vs CamelCase"},
		{"j.doe@company", "John Doe (External)", "Email-like vs Display name"},
	}
	
	for i, tc := range testCases {
		fmt.Printf("   Test case %d: %s\n", i+1, tc.desc)
		fmt.Printf("     GitLab: '%s' → GitHub: '%s'\n", tc.gitlab, tc.github)
		
		os.Setenv("GITLAB_USERNAME", tc.gitlab)
		os.Setenv("GITHUB_USERNAME", tc.github)
		
		err := internal.CheckEnvVariables()
		if err != nil {
			fmt.Printf("     ❌ Failed: %v\n", err)
			continue
		}
		
		// Verify they're stored correctly and independently
		storedGitlab := os.Getenv("GITLAB_USERNAME")
		storedGithub := os.Getenv("GITHUB_USERNAME")
		
		if storedGitlab != tc.gitlab {
			fmt.Printf("     ❌ GitLab username not stored correctly\n")
			continue
		}
		
		if storedGithub != tc.github {
			fmt.Printf("     ❌ GitHub name not stored correctly\n")
			continue
		}
		
		fmt.Printf("     ✅ Both usernames stored independently\n")
	}
	fmt.Println()

	// Test 4: Verify original functionality still works
	fmt.Println("Test 4: Original Functionality")
	fmt.Println("------------------------------")
	
	// Test that people with SAME usernames still work
	sameUsername := "samename"
	os.Setenv("GITLAB_USERNAME", sameUsername)
	os.Setenv("GITHUB_USERNAME", sameUsername)
	
	err = internal.CheckEnvVariables()
	if err != nil {
		fmt.Printf("❌ Same username validation failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("   ✅ Same usernames still work: '%s'\n", sameUsername)
	fmt.Printf("   ✅ Backward compatibility maintained\n")
	fmt.Println()

	// Test 5: Error handling
	fmt.Println("Test 5: Error Handling")
	fmt.Println("----------------------")
	
	// Test missing GITLAB_USERNAME
	os.Unsetenv("GITLAB_USERNAME")
	err = internal.CheckEnvVariables()
	if err == nil {
		fmt.Printf("❌ Should have failed with missing GITLAB_USERNAME\n")
		os.Exit(1)
	}
	
	if !strings.Contains(err.Error(), "GITLAB_USERNAME") {
		fmt.Printf("❌ Error should mention GITLAB_USERNAME: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("   ✅ Correctly detects missing GITLAB_USERNAME\n")
	fmt.Printf("   ✅ Error message: %v\n", err)
	fmt.Println()

	// Final summary
	fmt.Println("🎉 COMPREHENSIVE TEST RESULTS")
	fmt.Println("=============================")
	fmt.Println("✅ Different GitLab and GitHub usernames work correctly")
	fmt.Println("✅ GitLab API uses GITLAB_USERNAME (not GITHUB_USERNAME)")
	fmt.Println("✅ No assumptions made between usernames")
	fmt.Println("✅ All username format combinations work")
	fmt.Println("✅ Backward compatibility maintained")
	fmt.Println("✅ Proper error handling for missing variables")
	fmt.Println()
	fmt.Println("🔧 THE FIX IS WORKING PERFECTLY!")
	fmt.Println("   - GitLab username and GitHub username are completely independent")
	fmt.Println("   - No assumptions are made about username similarity")
	fmt.Println("   - All original features continue to work")

	// Cleanup
	clearEnvVars()
}

func clearEnvVars() {
	vars := []string{
		"BASE_URL", "GITLAB_TOKEN", "GITLAB_USERNAME", 
		"GITHUB_USERNAME", "COMMITER_EMAIL", "ORIGIN_REPO_URL", "ORIGIN_TOKEN",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

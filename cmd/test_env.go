package main

import (
	"fmt"
	"os"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

func main() {
	fmt.Println("🧪 Testing Environment Variables...")
	
	err := internal.CheckEnvVariables()
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ All environment variables are valid!")
	
	// Show the key difference
	gitlabUser := os.Getenv("GITLAB_USERNAME")
	githubUser := os.Getenv("GITHUB_USERNAME")
	
	fmt.Printf("GitLab Username: %s\n", gitlabUser)
	fmt.Printf("GitHub Username: %s\n", githubUser)
	
	if gitlabUser != githubUser {
		fmt.Println("🎉 Successfully supports different usernames!")
	} else {
		fmt.Println("ℹ️  Usernames are the same (which is also fine)")
	}
}

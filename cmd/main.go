package main

import (
	"log"

	"sync"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/furmanp/gitlab-activity-importer/internal/services"
)

func main() {
	startNow := time.Now()

	err := internal.LoadEnv()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config, err := internal.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	gitlabUser, err := services.GetGitlabUser(config.BaseURL)

	if err != nil {
		log.Fatalf("Error during reading GitLab User data: %v", err)
	}

	gitLabUserID := gitlabUser.ID

	projectIds, err := services.GetUsersProjectsIds(gitLabUserID, config.BaseURL, config.GitlabToken)

	if err != nil {
		log.Fatalf("Error during getting users projects: %v", err)
	}
	if len(projectIds) == 0 {
		log.Print("No contributions found for this user. Closing the program.")
		return
	}

	log.Printf("Found contributions in %d projects", len(projectIds))

	repo := services.OpenOrInitClone()

	err = services.PullLatestChanges(repo, config)
	if err != nil {
		log.Fatalf("Error pulling latest changes: %v", err)
	}

	commitChannel := make(chan []internal.Commit, len(projectIds))

	var wg sync.WaitGroup
	wg.Add(1)

	var totalCommitsCreated int
	go func() {
		defer wg.Done()
		totalCommits := 0
		for commits := range commitChannel {
			if localCommits, err := services.CreateLocalCommit(repo, commits, config); err == nil {
				totalCommits += localCommits
			} else {
				log.Printf("Error creating local commit: %v", err)
			}
		}
		totalCommitsCreated = totalCommits
		log.Printf("Imported %v commits.\n", totalCommits)
	}()

	services.FetchAllCommits(projectIds, config.GitlabUsername, commitChannel, config.BaseURL, config.GitlabToken)

	wg.Wait()

	if totalCommitsCreated > 0 {
		if err := services.PushLocalCommits(repo, config); err != nil {
			log.Fatalf("Error pushing local commits: %v", err)
			return
		}
		log.Println("Successfully pushed commits to remote repository.")
	} else {
		log.Println("No new commits were created, skipping push operation.")
	}
	log.Printf("Operation took: %v in total.", time.Since(startNow))
}

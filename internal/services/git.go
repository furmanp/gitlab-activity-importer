package services

import (
	"fmt"
	"log"
	"os"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func OpenOrInitClone() *git.Repository {
	repoPath := internal.GetHomeDirectory() + "/commits-importer/"

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			log.Println("Repository doesn't exist. Cloning new repository from remote.")
			repo, err = cloneRemoteRepo()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("Failed to open or initialize the repository:", err)
		}
	} else {
		log.Println("Opened existing repository.")
	}
	return repo
}

func cloneRemoteRepo() (*git.Repository, error) {
	homeDir := internal.GetHomeDirectory() + "/commits-importer/"
	repoURL := os.Getenv("ORIGIN_REPO_URL")

	repo, err := git.PlainClone(homeDir, false, &git.CloneOptions{
		URL: repoURL,
		Auth: &http.BasicAuth{
			Username: os.Getenv("GITHUB_USERNAME"),
			Password: os.Getenv("ORIGIN_TOKEN"),
		},
		Progress: os.Stdout,
	})

	if err != nil {
		if err == transport.ErrEmptyRemoteRepository {
			newRepo, initErr := git.PlainInit(homeDir, false)
			if initErr != nil {
				_ = os.RemoveAll(homeDir)
				return nil, initErr
			}

			_, remoteErr := newRepo.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{repoURL},
			})
			if remoteErr != nil {
				return nil, remoteErr
			}

			return newRepo, nil
		}
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	return repo, nil
}

func CreateLocalCommit(repo *git.Repository, commits []internal.Commit) (int, error) {
	if len(commits) == 0 {
		log.Println("No commits to process")
		return 0, nil
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return 0, fmt.Errorf("failed to get worktree: %w", err)
	}

	repoPath := internal.GetHomeDirectory() + "/commits-importer/"
	filePath := repoPath + "/readme.md"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to create file: %w", err)
		}
		_, err = file.WriteString("Just a readme.")
		if err != nil {
			file.Close()
			return 0, fmt.Errorf("failed to write to file: %w", err)
		}
		if err := file.Close(); err != nil {
			return 0, fmt.Errorf("failed to close file: %w", err)
		}
	}

	_, err = workTree.Add("readme.md")
	if err != nil {
		return 0, fmt.Errorf("failed to add readme.md to index: %w", err)
	}

	existingCommitSet, err := getAllExistingCommitSHAs(repo)
	if err != nil {
		return 0, fmt.Errorf("failed to get existing commits: %w", err)
	}

	totalCommits := 0
	for _, commit := range commits {
		if !existingCommitSet[commit.ID] {
			newCommit, err := workTree.Commit(commit.ID, &git.CommitOptions{
				Author: &object.Signature{
					Name:  os.Getenv("GITHUB_USERNAME"),
					Email: os.Getenv("COMMITER_EMAIL"),
					When:  commit.AuthoredDate,
				},
				Committer: &object.Signature{
					Name:  os.Getenv("GITHUB_USERNAME"),
					Email: os.Getenv("COMMITER_EMAIL"),
					When:  commit.AuthoredDate,
				},
				AllowEmptyCommits: true,
			})
			if err != nil {
				return 0, fmt.Errorf("failed to create commit %s: %w", commit.ID, err)
			}

			obj, err := repo.CommitObject(newCommit)
			if err != nil {
				return 0, fmt.Errorf("failed to get commit object for %s: %w", newCommit, err)
			}

			log.Printf("Created commit: %s\n", obj.Hash)
			totalCommits++
		} else {
			log.Printf("Commit: %v is already imported \n", commit.ID)
		}
	}
	return totalCommits, nil
}

func getAllExistingCommitSHAs(repo *git.Repository) (map[string]bool, error) {
	existingCommits := make(map[string]bool)
	ref, err := repo.Reference("HEAD", true)
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return existingCommits, nil
		}
		return nil, fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %v", err)
	}
	defer iter.Close()

	err = iter.ForEach(func(c *object.Commit) error {
		existingCommits[c.Message] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %v", err)
	}

	return existingCommits, nil
}

func PushLocalCommits(repo *git.Repository) {
	err := repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: os.Getenv("GITHUB_USERNAME"),
			Password: os.Getenv("ORIGIN_TOKEN"),
		},
		Progress: os.Stdout,
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Println("No changes to push, everything is up to date.")
		} else {
			log.Fatalf("Error pushing to Github: %v", err)
		}
	}
}

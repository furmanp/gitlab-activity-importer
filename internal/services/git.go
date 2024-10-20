package services

import (
	"fmt"
	"log"
	"os"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func OpenOrInitRepo() *git.Repository {
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
			Username: os.Getenv("COMMITER_NAME"),
			Password: os.Getenv("ORIGIN_TOKEN"),
		},
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatalf("Error cloning the repository: %v", err)
		return nil, err
	}
	return repo, nil
}

func CreateLocalCommit(repo *git.Repository, commit []internal.Commit) int {
	workTree, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	repoPath := internal.GetHomeDirectory() + "/commits-importer/"
	filePath := repoPath + "/readme.md"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString("Just a readme.")
		file.Close()
	}

	_, err = workTree.Add("readme.md")
	if err != nil {
		log.Fatal(err)
	}

	totalCommits := 0
	for index := range commit {
		isDuplicate, _ := checkIfCommitExists(repo, commit[index])

		if !isDuplicate {
			newCommit, err := workTree.Commit(commit[index].ID, &git.CommitOptions{
				Author: &object.Signature{
					Name:  os.Getenv("COMMITER_NAME"),
					Email: os.Getenv("COMMITER_EMAIL"),
					When:  commit[index].AuthoredDate,
				},
				Committer: &object.Signature{
					Name:  os.Getenv("COMMITER_NAME"),
					Email: os.Getenv("COMMITER_EMAIL"),
					When:  commit[index].AuthoredDate,
				},
			})
			if err != nil {
				log.Fatal(err)
			}

			obj, err := repo.CommitObject(newCommit)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Created commit: %s\n", obj.Hash)
			totalCommits += 1
		} else {
			log.Printf("Commit: %v is already imported \n", commit[index].ID)
		}
	}
	return totalCommits
}

func checkIfCommitExists(repo *git.Repository, commit internal.Commit) (bool, error) {
	ref, err := repo.Reference("HEAD", true)
	if err != nil {
		return false, err
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return false, err
	}

	err = iter.ForEach(func(c *object.Commit) error {
		if c.Message == commit.ID {
			return fmt.Errorf("duplicate commit found")
		}
		return nil
	})

	if err != nil {
		return true, err
	}

	return false, nil
}

func PushLocalCommits(repo *git.Repository) {
	err := repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: os.Getenv("COMMITER_NAME"),
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

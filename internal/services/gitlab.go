package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

func GetGitlabUser() (internal.GitLabUser, error) {
	url := os.Getenv("BASE_URL")

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%v/api/v4/user", url), nil)
	if err != nil {
		return internal.GitLabUser{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))

	res, err := client.Do(req)
	if err != nil {
		return internal.GitLabUser{}, fmt.Errorf("error making the request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return internal.GitLabUser{}, fmt.Errorf("status %d: %s", res.StatusCode, string(body))
	}

	var user internal.GitLabUser
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return internal.GitLabUser{}, fmt.Errorf("decode error: %w", err)
	}

	return user, nil
}

func GetUsersProjectsIds(userId int) ([]int, error) {
	base := os.Getenv("BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")

	allProjectIds := make([]int, 0, 128)
	client := &http.Client{Timeout: 30 * time.Second}

	for page := 1; ; {
		req, err := http.NewRequestWithContext(context.Background(),
			"GET",
			fmt.Sprintf("%s/api/v4/users/%d/contributed_projects?per_page=100&page=%d", base, userId, page),
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("PRIVATE-TOKEN", token)

		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		var next string
		func() {
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(res.Body)
				err = fmt.Errorf("request failed with status code: %d: %s", res.StatusCode, string(body))
				return
			}

			var projects []struct {
				ID int `json:"id"`
			}
			if derr := json.NewDecoder(res.Body).Decode(&projects); derr != nil {
				err = fmt.Errorf("error parsing JSON: %w", derr)
				return
			}

			for _, p := range projects {
				allProjectIds = append(allProjectIds, p.ID)
			}

			next = res.Header.Get("X-Next-Page")
		}()
		if err != nil {
			return nil, err
		}

		if next == "" {
			break
		}
		n, convErr := strconv.Atoi(next)
		if convErr != nil || n <= page {
			break
		}
		page = n
	}

	return allProjectIds, nil
}

func GetProjectCommits(projectId int, gitlabUserName string) ([]internal.Commit, error) {
	base := os.Getenv("BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")

	var allCommits []internal.Commit
	client := &http.Client{Timeout: 30 * time.Second}
	for page := 1; ; {
		req, err := http.NewRequestWithContext(context.Background(), "GET",
			fmt.Sprintf("%s/api/v4/projects/%d/repository/commits?author=%s&per_page=100&page=%d",
				base, projectId, url.QueryEscape(gitlabUserName), page), nil)
		if err != nil {
			return nil, fmt.Errorf("error fetching the commits: %w", err)
		}
		req.Header.Set("PRIVATE-TOKEN", token)

		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		var next string
		func() {
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(res.Body)
				err = fmt.Errorf("request failed with status code: %d: %s", res.StatusCode, string(body))
				return
			}
			var batch []internal.Commit
			if derr := json.NewDecoder(res.Body).Decode(&batch); derr != nil {
				err = fmt.Errorf("error parsing JSON: %w", derr)
				return
			}

			allCommits = append(allCommits, batch...)
			next = res.Header.Get("X-Next-Page")
		}()
		if err != nil {
			return nil, err
		}
		if next == "" {
			break
		}
		n, convErr := strconv.Atoi(next)
		if convErr != nil || n <= page {
			break
		}
		page = n
	}
	if len(allCommits) == 0 {
		return nil, fmt.Errorf("found no commits in project no.:%v", projectId)
	}
	return allCommits, nil
}

func FetchAllCommits(projectIds []int, gitlabUserName string, commitChannel chan []internal.Commit) {
	var wg sync.WaitGroup
	var validCommitsFound atomic.Bool

	for _, projectId := range projectIds {
		wg.Add(1)

		go func(projId int) {
			defer wg.Done()

			commits, err := GetProjectCommits(projId, gitlabUserName)
			if err != nil {
				log.Printf("Error fetching commits for project %d: %v", projId, err)
				return
			}
			if len(commits) > 0 {
				commitChannel <- commits
				validCommitsFound.Store(true)
			}

		}(projectId)
	}

	wg.Wait()

	if !validCommitsFound.Load() {
		log.Println("No valid commits found across any projects")
	}

	close(commitChannel)
}

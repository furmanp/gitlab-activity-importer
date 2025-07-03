package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

func GetGitlabUser() (internal.GitLabUser, error) {
	url := os.Getenv("BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/user", url), nil)
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
		return internal.GitLabUser{}, fmt.Errorf("request failed with status code: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return internal.GitLabUser{}, fmt.Errorf("error reading the response body: %v", err)
	}

	var user internal.GitLabUser
	if err := json.Unmarshal(body, &user); err != nil {
		return internal.GitLabUser{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	return user, nil
}

func GetUsersProjectsIds(userId int) ([]int, error) {
	url := os.Getenv("BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/users/%v/contributed_projects", url, userId), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making the request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var projects []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	projectIds := make([]int, len(projects))
	for i, project := range projects {
		projectIds[i] = project.ID
	}

	return projectIds, nil
}

func GetProjectCommits(projectId int, userName string) ([]internal.Commit, error) {
	url := os.Getenv("BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")

	var allCommits []internal.Commit
	client := &http.Client{}
	page := 1

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/projects/%v/repository/commits?author=%v&per_page=100&page=%d", url, projectId, userName, page), nil)
		if err != nil {
			return nil, fmt.Errorf("error fetching the commits: %v", err)
		}

		req.Header.Set("PRIVATE-TOKEN", token)
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making the request: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status code: %v", res.StatusCode)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading the response body: %v", err)
		}

		var commits []internal.Commit
		err = json.Unmarshal(body, &commits)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}

		if len(commits) == 0 {
			break
		}

		allCommits = append(allCommits, commits...)

		page++
	}

	if len(allCommits) == 0 {
		return nil, fmt.Errorf("found no commits in project no.:%v", projectId)
	}

	log.Printf("Found total of %v commits in project no.:%v \n", len(allCommits), projectId)

	return allCommits, nil
}

func FetchAllCommits(projectIds []int, commiterName string, commitChannel chan []internal.Commit) {
	var wg sync.WaitGroup
	validCommitsFound := false

	for _, projectId := range projectIds {
		wg.Add(1)

		go func(projId int) {
			defer wg.Done()

			commits, err := GetProjectCommits(projId, commiterName)
			if err != nil {
				log.Printf("Error fetching commits for project %d: %v", projId, err)
				return
			}
			if len(commits) > 0 {
				commitChannel <- commits
				validCommitsFound = true
			}

		}(projectId)
	}

	wg.Wait()

	if !validCommitsFound {
		log.Println("No valid commits found across any projects")
	}

	close(commitChannel)
}

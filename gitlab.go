package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func getGitlabUser() string {
	url := os.Getenv("BASE_URL")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/user", url), nil)
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))

	res, err := client.Do(req)

	if err != nil {
		fmt.Print("something went wrong with your request", err)
	}

	if res.StatusCode == http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal("something went wrong")
		}
		json := string(body)

		return json
	}

	return "User not found"
}

func getUsersProjectsIds(userId int) ([]int, error) {
	url := os.Getenv("BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/users/%v/contributed_projects", url, userId), nil)
	if err != nil {
		log.Fatalf("Error creating the request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading the response body: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	if len(result) == 0 {
		log.Fatalf("No contributed projects found")
	}

	var projectIds []int
	for index := range result {
		id := result[index]["id"].(float64)
		projectIds = append(projectIds, int(id))
	}

	return projectIds, nil
}

func getProjectCommits(projectId int, userName string) ([]Commit, error) {
	url := os.Getenv("BASE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v4/projects/%v/repository/commits?author=%v&per_page=100&page=1", url, projectId, userName), nil)
	if err != nil {
		log.Fatalf("Error fetching the commits: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading the response body: %v", err)
	}

	var commits []Commit

	err = json.Unmarshal([]byte(body), &commits)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits for project found")

	}

	log.Printf("Found %v commits in this project \n", len(commits))
	return commits, nil

}

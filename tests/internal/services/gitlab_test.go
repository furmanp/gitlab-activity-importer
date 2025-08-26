package services_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/furmanp/gitlab-activity-importer/internal/services"
)

func TestGetGitlabUser(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		statusCode   int
		expectError  bool
		expectedUser internal.GitLabUser
	}{
		{
			name:        "valid token and successful response",
			token:       "valid-token",
			statusCode:  200,
			expectError: false,
			expectedUser: internal.GitLabUser{
				ID:       1,
				Username: "testuser",
			},
		},
		{
			name:         "missing token",
			token:        "",
			statusCode:   401,
			expectError:  true,
			expectedUser: internal.GitLabUser{},
		},
		{
			name:         "invalid token",
			token:        "invalid-token",
			statusCode:   401,
			expectError:  true,
			expectedUser: internal.GitLabUser{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.token != "" {
				os.Setenv("GITLAB_TOKEN", tt.token)
			} else {
				os.Unsetenv("GITLAB_TOKEN")
			}
			defer os.Unsetenv("GITLAB_TOKEN")

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if r.Header.Get("PRIVATE-TOKEN") != tt.token {
					t.Errorf("Expected PRIVATE-TOKEN '%s', got '%s'", tt.token, r.Header.Get("PRIVATE-TOKEN"))
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					fmt.Fprint(w, `{"username":"testuser","id":1}`)
				} else {
					fmt.Fprint(w, "Unauthorized")
				}
			}))
			defer mockServer.Close()

			os.Setenv("BASE_URL", mockServer.URL)
			defer os.Unsetenv("BASE_URL")

			result, err := services.GetGitlabUser()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetGitlabUser returned error: %v", err)
			}

			if result != tt.expectedUser {
				t.Errorf("Expected user '%v', got '%v'", tt.expectedUser, result)
			}
		})
	}
}
func TestGetUsersProjectsIds(t *testing.T) {
	tests := []struct {
		name             string
		userId           int
		statusCode       int
		expectedResponse []map[string]interface{}
		expectedIds      []int
		expectError      bool
	}{
		{
			name:       "projects found",
			userId:     1,
			statusCode: 200,
			expectedResponse: []map[string]interface{}{
				{"id": 1, "name": "Project1"},
				{"id": 2, "name": "Project2"},
			},
			expectedIds: []int{1, 2},
			expectError: false,
		},
		{
			name:             "no projects found",
			userId:           2,
			statusCode:       200,
			expectedResponse: []map[string]interface{}{},
			expectedIds:      []int{},
			expectError:      false,
		},
		{
			name:       "user not found",
			userId:     2,
			statusCode: 404,
			expectedResponse: []map[string]interface{}{
				{"message": "404 User Not Found"},
			},
			expectedIds: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BASE_URL", "http://test-url.com")
			os.Setenv("GITLAB_TOKEN", "test-token")
			defer os.Unsetenv("BASE_URL")
			defer os.Unsetenv("GITLAB_TOKEN")

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedURL := fmt.Sprintf("/api/v4/users/%d/contributed_projects", tt.userId)
				if r.URL.Path != expectedURL {
					t.Errorf("Expected URL '%s', got '%s'", expectedURL, r.URL.Path)
				}
				if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
					t.Errorf("Expected PRIVATE-TOKEN 'test-token', got '%s'", r.Header.Get("PRIVATE-TOKEN"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				responseData, err := json.Marshal(tt.expectedResponse)
				if err != nil {
					t.Fatalf("Failed to marshal response data: %v", err)
				}
				w.Write(responseData)
			}))
			defer mockServer.Close()

			os.Setenv("BASE_URL", mockServer.URL)

			result, err := services.GetUsersProjectsIds(tt.userId)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUsersProjectsIds returned error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expectedIds) {
				t.Errorf("Expected '%v', got '%v'", tt.expectedIds, result)
			}
		})
	}
}

func TestGetProjectsCommits(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		userName       string
		projectId      int
		responses      [][]internal.Commit
		statusCodes    []int
		expectedResult []internal.Commit
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:      "successful request with pagination",
			userName:  "user",
			projectId: 1,
			responses: [][]internal.Commit{
				{
					{
						ID:           "123",
						Message:      "first commit",
						AuthorName:   "John Doe",
						AuthorMail:   "john@doe.com",
						AuthoredDate: fixedTime,
					},
					{
						ID:           "456",
						Message:      "second commit",
						AuthorName:   "John Doe",
						AuthorMail:   "john@doe.com",
						AuthoredDate: fixedTime,
					},
				},
				{
					{
						ID:           "789",
						Message:      "third commit",
						AuthorName:   "John Doe",
						AuthorMail:   "john@doe.com",
						AuthoredDate: fixedTime,
					},
				},
				{},
			},
			statusCodes: []int{200, 200, 200},
			expectedResult: []internal.Commit{
				{
					ID:           "123",
					Message:      "first commit",
					AuthorName:   "John Doe",
					AuthorMail:   "john@doe.com",
					AuthoredDate: fixedTime,
				},
				{
					ID:           "456",
					Message:      "second commit",
					AuthorName:   "John Doe",
					AuthorMail:   "john@doe.com",
					AuthoredDate: fixedTime,
				},
				{
					ID:           "789",
					Message:      "third commit",
					AuthorName:   "John Doe",
					AuthorMail:   "john@doe.com",
					AuthoredDate: fixedTime,
				},
			},
			expectError: false,
		},
		{
			name:      "no commits found",
			userName:  "user",
			projectId: 2,
			responses: [][]internal.Commit{
				{},
			},
			statusCodes:    []int{200},
			expectError:    true,
			expectedErrMsg: "found no commits in project no.:2",
		},
		{
			name:           "unauthorized request",
			userName:       "user",
			projectId:      3,
			responses:      [][]internal.Commit{nil},
			statusCodes:    []int{401},
			expectError:    true,
			expectedErrMsg: "request failed with status code: 401: null",
		},
		{
			name:           "invalid json response",
			userName:       "user",
			projectId:      4,
			responses:      [][]internal.Commit{nil},
			statusCodes:    []int{200},
			expectError:    true,
			expectedErrMsg: "error parsing JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BASE_URL", "http://test-url.com")
			os.Setenv("GITLAB_TOKEN", "test-token")
			defer os.Unsetenv("BASE_URL")
			defer os.Unsetenv("GITLAB_TOKEN")

			requestCount := 0

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
					t.Errorf("Expected PRIVATE-TOKEN 'test-token', got '%s'", r.Header.Get("PRIVATE-TOKEN"))
				}

				expectedPath := fmt.Sprintf("/api/v4/projects/%d/repository/commits", tt.projectId)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				queryParams := r.URL.Query()
				if queryParams.Get("author") != tt.userName {
					t.Errorf("Expected author %s, got %s", tt.userName, queryParams.Get("author"))
				}
				if queryParams.Get("per_page") != "100" {
					t.Errorf("Expected per_page 100, got %s", queryParams.Get("per_page"))
				}

				pageNum := requestCount + 1
				if queryParams.Get("page") != fmt.Sprintf("%d", pageNum) {
					t.Errorf("Expected page %d, got %s", pageNum, queryParams.Get("page"))
				}

				w.Header().Set("Content-Type", "application/json")

				if requestCount+1 < len(tt.statusCodes) {
					w.Header().Set("X-Next-Page", fmt.Sprintf("%d", requestCount+2))
				} else {
					w.Header().Set("X-Next-Page", "")
				}

				if requestCount >= len(tt.statusCodes) {
					t.Fatalf("More requests than expected status codes")
				}
				w.WriteHeader(tt.statusCodes[requestCount])

				var responseBody []byte
				var err error

				if tt.name == "invalid json response" {
					responseBody = []byte(`{"invalid json`)
				} else if requestCount < len(tt.responses) {
					responseBody, err = json.Marshal(tt.responses[requestCount])
					if err != nil {
						t.Fatalf("Failed to marshal response data: %v", err)
					}
				}

				w.Write(responseBody)
				requestCount++
			}))
			defer mockServer.Close()

			os.Setenv("BASE_URL", mockServer.URL)

			result, err := services.GetProjectCommits(tt.projectId, tt.userName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedErrMsg, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Expected response %+v, got %+v", tt.expectedResult, result)
			}

			expectedRequests := len(tt.responses)
			if requestCount != expectedRequests {
				t.Errorf("Expected %d requests, got %d", expectedRequests, requestCount)
			}
		})
	}
}

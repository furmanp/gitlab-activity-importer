package internal_test

import (
	"strings"
	"testing"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

var allEnvVars = []string{
	"BASE_URL",
	"GITLAB_TOKEN",
	"GITLAB_USERNAME",
	"GH_USERNAME",
	"COMMITER_EMAIL",
	"ORIGIN_REPO_URL",
	"ORIGIN_TOKEN",
}

func fullEnv() map[string]string {
	return map[string]string{
		"BASE_URL":        "http://base-url-value.com",
		"GITLAB_TOKEN":    "gitlab-token-value",
		"GITLAB_USERNAME": "gitlab-username-value",
		"GH_USERNAME":     "gh-username-value",
		"COMMITER_EMAIL":  "commiter-email-value",
		"ORIGIN_REPO_URL": "http://origin-repo-url-value.com",
		"ORIGIN_TOKEN":    "origin-token-value",
	}
}

func applyEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for _, name := range allEnvVars {
		t.Setenv(name, env[name])
	}
}

func TestLoadMapsEveryVariableToItsField(t *testing.T) {
	env := fullEnv()
	applyEnv(t, env)

	got, err := internal.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := internal.Config{
		BaseURL:        env["BASE_URL"],
		GitlabUsername: env["GITLAB_USERNAME"],
		GitlabToken:    env["GITLAB_TOKEN"],
		GithubUsername: env["GH_USERNAME"],
		CommitterEmail: env["COMMITER_EMAIL"],
		OriginRepoURL:  env["ORIGIN_REPO_URL"],
		OriginToken:    env["ORIGIN_TOKEN"],
	}

	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadReportsMissingVariables(t *testing.T) {
	tests := []struct {
		name    string
		missing []string
	}{
		{
			name: "all required variables set",
		},
		{
			name:    "missing one variable",
			missing: []string{"ORIGIN_REPO_URL"},
		},
		{
			name:    "missing multiple variables",
			missing: []string{"GITLAB_TOKEN", "GITLAB_USERNAME", "GH_USERNAME", "COMMITER_EMAIL", "ORIGIN_REPO_URL", "ORIGIN_TOKEN"},
		},
		{
			name:    "no variables set",
			missing: allEnvVars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := make(map[string]bool, len(tt.missing))
			env := fullEnv()
			for _, name := range tt.missing {
				missing[name] = true
				delete(env, name)
			}
			applyEnv(t, env)

			got, err := internal.Load()

			if len(tt.missing) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error but got none")
			}

			for _, name := range allEnvVars {
				named := strings.Contains(err.Error(), name)
				if missing[name] && !named {
					t.Errorf("error should name the missing %s, got: %v", name, err)
				}
				if !missing[name] && named {
					t.Errorf("error names %s, which was set: %v", name, err)
				}
			}

			if got != (internal.Config{}) {
				t.Errorf("expected a zero Config alongside the error, got %+v", got)
			}
		})
	}
}

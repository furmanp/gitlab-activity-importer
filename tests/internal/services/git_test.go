package services_test

import (
	"testing"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/furmanp/gitlab-activity-importer/internal/services"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func setupEmptyRemote(t *testing.T) (string, internal.Config) {
	t.Helper()

	root := t.TempDir()
	remotePath := root + "/remote.git"

	if _, err := git.PlainInit(remotePath, true); err != nil {
		t.Fatalf("failed to create bare remote: %v", err)
	}

	cfg := internal.Config{
		GithubUsername: "test-user",
		CommitterEmail: "test-user@example.com",
		OriginRepoURL:  remotePath,
		OriginToken:    "test-token",
	}

	t.Setenv("HOME", root+"/home")
	t.Setenv("USERPROFILE", root+"/home")
	t.Setenv("ORIGIN_REPO_URL", cfg.OriginRepoURL)
	t.Setenv("GH_USERNAME", cfg.GithubUsername)
	t.Setenv("COMMITER_EMAIL", cfg.CommitterEmail)
	t.Setenv("ORIGIN_TOKEN", cfg.OriginToken)

	return remotePath, cfg
}

func TestFirstRunAgainstEmptyRemote(t *testing.T) {
	_, cfg := setupEmptyRemote(t)

	repo := services.OpenOrInitClone()

	if err := services.PullLatestChanges(repo, cfg); err != nil {
		t.Fatalf("PullLatestChanges against an empty remote should be a no-op, got: %v", err)
	}

	head, err := repo.Reference(plumbing.HEAD, false)
	if err != nil {
		t.Fatalf("failed to read HEAD: %v", err)
	}

	if got, want := head.Target().String(), "refs/heads/main"; got != want {
		t.Errorf("HEAD = %q, want %q (the contribution graph only counts the default branch)", got, want)
	}
}

func TestImportAndPushToEmptyRemote(t *testing.T) {
	remotePath, cfg := setupEmptyRemote(t)

	repo := services.OpenOrInitClone()
	if err := services.PullLatestChanges(repo, cfg); err != nil {
		t.Fatalf("PullLatestChanges: %v", err)
	}

	commits := []internal.Commit{
		{ID: "aaa111", AuthoredDate: time.Date(2025, 3, 1, 9, 0, 0, 0, time.UTC)},
		{ID: "bbb222", AuthoredDate: time.Date(2025, 3, 2, 9, 0, 0, 0, time.UTC)},
	}

	created, err := services.CreateLocalCommit(repo, commits, cfg)
	if err != nil {
		t.Fatalf("CreateLocalCommit: %v", err)
	}
	if created != len(commits) {
		t.Fatalf("created %d commits, want %d", created, len(commits))
	}

	if err := services.PushLocalCommits(repo, cfg); err != nil {
		t.Fatalf("PushLocalCommits: %v", err)
	}

	remote, err := git.PlainOpen(remotePath)
	if err != nil {
		t.Fatalf("failed to open remote: %v", err)
	}

	ref, err := remote.Reference(plumbing.NewBranchReferenceName("main"), false)
	if err != nil {
		t.Fatalf("expected branch main on the remote: %v", err)
	}

	iter, err := remote.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		t.Fatalf("failed to read remote log: %v", err)
	}
	defer iter.Close()

	pushed := map[string]bool{}
	if err := iter.ForEach(func(c *object.Commit) error {
		pushed[c.Message] = true
		return nil
	}); err != nil {
		t.Fatalf("failed to iterate remote log: %v", err)
	}

	for _, c := range commits {
		if !pushed[c.ID] {
			t.Errorf("commit %s missing from the remote", c.ID)
		}
	}
}

func TestSecondRunIsIdempotent(t *testing.T) {
	_, cfg := setupEmptyRemote(t)

	commits := []internal.Commit{
		{ID: "aaa111", AuthoredDate: time.Date(2025, 3, 1, 9, 0, 0, 0, time.UTC)},
		{ID: "bbb222", AuthoredDate: time.Date(2025, 3, 2, 9, 0, 0, 0, time.UTC)},
	}

	first := services.OpenOrInitClone()
	if err := services.PullLatestChanges(first, cfg); err != nil {
		t.Fatalf("first run, PullLatestChanges: %v", err)
	}
	created, err := services.CreateLocalCommit(first, commits, cfg)
	if err != nil {
		t.Fatalf("first run, CreateLocalCommit: %v", err)
	}
	if created != len(commits) {
		t.Fatalf("first run created %d commits, want %d", created, len(commits))
	}
	if err := services.PushLocalCommits(first, cfg); err != nil {
		t.Fatalf("first run, PushLocalCommits: %v", err)
	}

	second := services.OpenOrInitClone()
	if err := services.PullLatestChanges(second, cfg); err != nil {
		t.Fatalf("second run, PullLatestChanges: %v", err)
	}
	created, err = services.CreateLocalCommit(second, commits, cfg)
	if err != nil {
		t.Fatalf("second run, CreateLocalCommit: %v", err)
	}
	if created != 0 {
		t.Errorf("second run created %d commits, want 0", created)
	}
}

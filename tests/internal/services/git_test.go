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

func setupEmptyRemote(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	remotePath := root + "/remote.git"

	if _, err := git.PlainInit(remotePath, true); err != nil {
		t.Fatalf("failed to create bare remote: %v", err)
	}

	t.Setenv("HOME", root+"/home")
	t.Setenv("USERPROFILE", root+"/home")
	t.Setenv("ORIGIN_REPO_URL", remotePath)
	t.Setenv("GH_USERNAME", "test-user")
	t.Setenv("COMMITER_EMAIL", "test-user@example.com")
	t.Setenv("ORIGIN_TOKEN", "test-token")

	return remotePath
}

func TestFirstRunAgainstEmptyRemote(t *testing.T) {
	setupEmptyRemote(t)

	repo := services.OpenOrInitClone()

	if err := services.PullLatestChanges(repo); err != nil {
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
	remotePath := setupEmptyRemote(t)

	repo := services.OpenOrInitClone()
	if err := services.PullLatestChanges(repo); err != nil {
		t.Fatalf("PullLatestChanges: %v", err)
	}

	commits := []internal.Commit{
		{ID: "aaa111", AuthoredDate: time.Date(2025, 3, 1, 9, 0, 0, 0, time.UTC)},
		{ID: "bbb222", AuthoredDate: time.Date(2025, 3, 2, 9, 0, 0, 0, time.UTC)},
	}

	created, err := services.CreateLocalCommit(repo, commits)
	if err != nil {
		t.Fatalf("CreateLocalCommit: %v", err)
	}
	if created != len(commits) {
		t.Fatalf("created %d commits, want %d", created, len(commits))
	}

	if err := services.PushLocalCommits(repo); err != nil {
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
	setupEmptyRemote(t)

	repo := services.OpenOrInitClone()
	if err := services.PullLatestChanges(repo); err != nil {
		t.Fatalf("PullLatestChanges: %v", err)
	}

	commits := []internal.Commit{
		{ID: "aaa111", AuthoredDate: time.Date(2025, 3, 1, 9, 0, 0, 0, time.UTC)},
	}

	if _, err := services.CreateLocalCommit(repo, commits); err != nil {
		t.Fatalf("first import: %v", err)
	}

	created, err := services.CreateLocalCommit(repo, commits)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if created != 0 {
		t.Errorf("second import created %d commits, want 0", created)
	}
}

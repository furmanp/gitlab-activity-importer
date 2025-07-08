// Package gitlab_activity_importer provides a tool to transfer your GitLab commit history
// to GitHub, reflecting your GitLab activity on GitHub’s contribution graph.
//
// # Overview
//
// This tool fetches your commit history from private GitLab repositories and imports it
// into a specified GitHub repository, creating a visual representation of your activity
// on GitHub’s contribution graph. It can be configured for automated daily imports (via
// GitHub Actions) or manual runs.
//
// Features:
//
//   - Automated daily imports: Syncs your GitLab activity with GitHub automatically.
//   - Manual imports: Allows on-demand updates.
//   - Secure data handling: Uses repository secrets for configuration.
//
// # Usage
//
// See the project README for setup and configuration instructions:
// https://github.com/furmanp/gitlab-activity-importer
package gitlab_activity_importer

// Version is the current version of the tool.
var Version = "{{VERSION}}"

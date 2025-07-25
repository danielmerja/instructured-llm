package documentloaders

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewGitHubIssuesLoader(t *testing.T) {
	// Test creation without environment variable
	originalToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	os.Unsetenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", originalToken)
		}
	}()

	_, err := NewGitHubIssuesLoader("owner/repo")
	if err == nil {
		t.Error("Expected error when GITHUB_PERSONAL_ACCESS_TOKEN is not set")
	}

	// Test creation with empty repo
	_, err = NewGitHubIssuesLoader("")
	if err == nil {
		t.Error("Expected error when repository is empty")
	}

	// Test creation with token option
	loader, err := NewGitHubIssuesLoader("owner/repo", WithAccessToken("test-token"))
	if err != nil {
		t.Fatalf("Failed to create loader with access token: %v", err)
	}

	if loader.AccessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got '%s'", loader.AccessToken)
	}

	if loader.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", loader.Repo)
	}

	// Test default values
	if loader.IncludePRs != true {
		t.Error("Expected IncludePRs to default to true")
	}

	if loader.State != "open" {
		t.Errorf("Expected state to default to 'open', got '%s'", loader.State)
	}

	if loader.GitHubAPIURL != "https://api.github.com" {
		t.Errorf("Expected default GitHub API URL, got '%s'", loader.GitHubAPIURL)
	}
}

func TestGitHubIssuesLoaderOptions(t *testing.T) {
	loader, err := NewGitHubIssuesLoader("owner/repo",
		WithAccessToken("test-token"),
		WithIncludePRs(false),
		WithState("closed"),
		WithLabels([]string{"bug", "enhancement"}),
		WithMilestone("v1.0"),
		WithAssignee("testuser"),
		WithCreator("creator"),
		WithSort("updated", "desc"),
		WithSince("2023-01-01T00:00:00Z"),
		WithPagination(2, 50),
	)

	if err != nil {
		t.Fatalf("Failed to create loader with options: %v", err)
	}

	if loader.IncludePRs != false {
		t.Error("Expected IncludePRs to be false")
	}

	if loader.State != "closed" {
		t.Errorf("Expected state 'closed', got '%s'", loader.State)
	}

	if len(loader.Labels) != 2 || loader.Labels[0] != "bug" || loader.Labels[1] != "enhancement" {
		t.Errorf("Expected labels [bug, enhancement], got %v", loader.Labels)
	}

	if loader.Milestone == nil || *loader.Milestone != "v1.0" {
		t.Error("Expected milestone 'v1.0'")
	}

	if loader.Assignee != "testuser" {
		t.Errorf("Expected assignee 'testuser', got '%s'", loader.Assignee)
	}

	if loader.Creator != "creator" {
		t.Errorf("Expected creator 'creator', got '%s'", loader.Creator)
	}

	if loader.Sort != "updated" {
		t.Errorf("Expected sort 'updated', got '%s'", loader.Sort)
	}

	if loader.Direction != "desc" {
		t.Errorf("Expected direction 'desc', got '%s'", loader.Direction)
	}

	if loader.Since != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected since '2023-01-01T00:00:00Z', got '%s'", loader.Since)
	}

	if loader.Page == nil || *loader.Page != 2 {
		t.Error("Expected page 2")
	}

	if loader.PerPage == nil || *loader.PerPage != 50 {
		t.Error("Expected per_page 50")
	}
}

func TestGitHubIssuesLoaderBuildURL(t *testing.T) {
	loader, _ := NewGitHubIssuesLoader("owner/repo",
		WithAccessToken("test-token"),
		WithState("all"),
		WithLabels([]string{"bug", "enhancement"}),
		WithMilestone("v1.0"),
	)

	url := loader.buildURL()
	expectedBase := "https://api.github.com/repos/owner/repo/issues"

	if !strings.HasPrefix(url, expectedBase) {
		t.Errorf("Expected URL to start with '%s', got '%s'", expectedBase, url)
	}

	if !strings.Contains(url, "state=all") {
		t.Error("Expected URL to contain 'state=all'")
	}

	if !strings.Contains(url, "labels=bug%2Cenhancement") {
		t.Error("Expected URL to contain encoded labels")
	}

	if !strings.Contains(url, "milestone=v1.0") {
		t.Error("Expected URL to contain 'milestone=v1.0'")
	}
}

func TestNewGitHubFileLoader(t *testing.T) {
	// Test creation without environment variable
	originalToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	os.Unsetenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", originalToken)
		}
	}()

	_, err := NewGitHubFileLoader("owner/repo")
	if err == nil {
		t.Error("Expected error when GITHUB_PERSONAL_ACCESS_TOKEN is not set")
	}

	// Test creation with empty repo
	_, err = NewGitHubFileLoader("")
	if err == nil {
		t.Error("Expected error when repository is empty")
	}

	// Test creation with token option
	loader, err := NewGitHubFileLoader("owner/repo", WithFileAccessToken("test-token"))
	if err != nil {
		t.Fatalf("Failed to create file loader with access token: %v", err)
	}

	if loader.AccessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got '%s'", loader.AccessToken)
	}

	if loader.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", loader.Repo)
	}

	// Test default values
	if loader.Branch != "main" {
		t.Errorf("Expected branch to default to 'main', got '%s'", loader.Branch)
	}

	if loader.GitHubAPIURL != "https://api.github.com" {
		t.Errorf("Expected default GitHub API URL, got '%s'", loader.GitHubAPIURL)
	}
}

func TestGitHubFileLoaderOptions(t *testing.T) {
	filterFunc := func(path string) bool {
		return strings.HasSuffix(path, ".go")
	}

	loader, err := NewGitHubFileLoader("owner/repo",
		WithFileAccessToken("test-token"),
		WithBranch("develop"),
		WithFileFilter(filterFunc),
	)

	if err != nil {
		t.Fatalf("Failed to create file loader with options: %v", err)
	}

	if loader.Branch != "develop" {
		t.Errorf("Expected branch 'develop', got '%s'", loader.Branch)
	}

	if loader.FileFilter == nil {
		t.Error("Expected file filter to be set")
	}

	// Test the filter function
	if !loader.FileFilter("test.go") {
		t.Error("Expected filter to accept .go files")
	}

	if loader.FileFilter("test.txt") {
		t.Error("Expected filter to reject .txt files")
	}
}

func TestGitHubIssuesLoaderIntegration(t *testing.T) {
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: GITHUB_PERSONAL_ACCESS_TOKEN not set")
	}

	// Use a known public repository with issues
	loader, err := NewGitHubIssuesLoader("octocat/Hello-World",
		WithAccessToken(token),
		WithState("all"),
		WithPagination(1, 5), // Limit to avoid too many requests
	)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	docs, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load issues: %v", err)
	}

	if len(docs) == 0 {
		t.Log("No issues found (this might be expected for some repositories)")
		return
	}

	// Verify document structure
	doc := docs[0]
	if doc.Metadata == nil {
		t.Error("Expected metadata to be set")
	}

	expectedFields := []string{"url", "title", "creator", "state", "number", "is_pull_request"}
	for _, field := range expectedFields {
		if _, exists := doc.Metadata[field]; !exists {
			t.Errorf("Expected metadata field '%s' to exist", field)
		}
	}
}

func TestGitHubFileLoaderIntegration(t *testing.T) {
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: GITHUB_PERSONAL_ACCESS_TOKEN not set")
	}

	// Use a known public repository
	loader, err := NewGitHubFileLoader("octocat/Hello-World",
		WithFileAccessToken(token),
		WithFileFilter(func(path string) bool {
			// Only load README files to limit the test
			return strings.Contains(strings.ToLower(path), "readme")
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create file loader: %v", err)
	}

	docs, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load files: %v", err)
	}

	if len(docs) == 0 {
		t.Log("No README files found")
		return
	}

	// Verify document structure
	doc := docs[0]
	if doc.PageContent == "" {
		t.Error("Expected page content to be non-empty")
	}

	if doc.Metadata == nil {
		t.Error("Expected metadata to be set")
	}

	expectedFields := []string{"path", "sha", "source"}
	for _, field := range expectedFields {
		if _, exists := doc.Metadata[field]; !exists {
			t.Errorf("Expected metadata field '%s' to exist", field)
		}
	}
}

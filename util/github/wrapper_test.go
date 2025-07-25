package github

import (
	"os"
	"strings"
	"testing"

	githubapi "github.com/google/go-github/v74/github"
)

func TestNewGitHubAPIWrapper(t *testing.T) {
	// Test with empty config (should use environment variables)
	originalRepo := os.Getenv("GITHUB_REPOSITORY")
	originalAppID := os.Getenv("GITHUB_APP_ID")
	originalPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")

	// Clean up environment
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GITHUB_APP_ID")
	os.Unsetenv("GITHUB_APP_PRIVATE_KEY")

	defer func() {
		if originalRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalRepo)
		}
		if originalAppID != "" {
			os.Setenv("GITHUB_APP_ID", originalAppID)
		}
		if originalPrivateKey != "" {
			os.Setenv("GITHUB_APP_PRIVATE_KEY", originalPrivateKey)
		}
	}()

	// Test missing repository
	_, err := NewGitHubAPIWrapper(nil)
	if err == nil {
		t.Error("Expected error when GITHUB_REPOSITORY is missing")
	}
	if !strings.Contains(err.Error(), "GITHUB_REPOSITORY is required") {
		t.Errorf("Expected repository error, got: %v", err)
	}

	// Test missing app ID
	_, err = NewGitHubAPIWrapper(&Config{Repository: "owner/repo"})
	if err == nil {
		t.Error("Expected error when GITHUB_APP_ID is missing")
	}
	if !strings.Contains(err.Error(), "GITHUB_APP_ID is required") {
		t.Errorf("Expected app ID error, got: %v", err)
	}

	// Test missing private key
	_, err = NewGitHubAPIWrapper(&Config{
		Repository: "owner/repo",
		AppID:      "123456",
	})
	if err == nil {
		t.Error("Expected error when GITHUB_APP_PRIVATE_KEY is missing")
	}
	if !strings.Contains(err.Error(), "GITHUB_APP_PRIVATE_KEY is required") {
		t.Errorf("Expected private key error, got: %v", err)
	}

	// Test invalid repository format
	_, err = NewGitHubAPIWrapper(&Config{
		Repository: "invalid-repo-format",
		AppID:      "123456",
		PrivateKey: "fake-key",
	})
	if err == nil {
		t.Error("Expected error for invalid repository format")
	}
	if !strings.Contains(err.Error(), "invalid repository format") {
		t.Errorf("Expected repository format error, got: %v", err)
	}
}

func TestParseIssues(t *testing.T) {
	wrapper := &GitHubAPIWrapper{}

	// Test with nil slice
	result := wrapper.ParseIssues(nil)
	if len(result) != 0 {
		t.Error("Expected empty result for nil input")
	}

	// Note: We can't easily test with real GitHub issues without creating
	// complex mock objects, but we can test the structure
}

func TestParseIssuesEmpty(t *testing.T) {
	wrapper := &GitHubAPIWrapper{}

	result := wrapper.ParseIssues([]*githubapi.Issue{})
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}
}

func TestParsePullRequestsEmpty(t *testing.T) {
	wrapper := &GitHubAPIWrapper{}

	result := wrapper.ParsePullRequests([]*githubapi.PullRequest{})
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}
}

func TestRunInvalidMode(t *testing.T) {
	wrapper := &GitHubAPIWrapper{}

	_, err := wrapper.Run("invalid_mode", "test")
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("Expected invalid mode error, got: %v", err)
	}
}

func TestSetActiveBranch(t *testing.T) {
	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping test: GitHub credentials not available")
	}

	wrapper, err := NewGitHubAPIWrapper(&Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	// Test setting to non-existent branch
	result, err := wrapper.SetActiveBranch("non-existent-branch-12345")
	if err != nil {
		t.Errorf("SetActiveBranch should not return error, got: %v", err)
	}
	if !strings.Contains(result, "does not exist") {
		t.Errorf("Expected 'does not exist' message, got: %s", result)
	}
}

func TestRunModeDispatching(t *testing.T) {
	wrapper := &GitHubAPIWrapper{}

	// Test get_issue with invalid number
	_, err := wrapper.Run("get_issue", "invalid")
	if err == nil {
		t.Error("Expected error for invalid issue number")
	}

	// Test get_pull_request with invalid number
	_, err = wrapper.Run("get_pull_request", "invalid")
	if err == nil {
		t.Error("Expected error for invalid PR number")
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	// Test that config properly reads from environment
	originalRepo := os.Getenv("GITHUB_REPOSITORY")
	originalAppID := os.Getenv("GITHUB_APP_ID")
	originalPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")

	defer func() {
		if originalRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
		if originalAppID != "" {
			os.Setenv("GITHUB_APP_ID", originalAppID)
		} else {
			os.Unsetenv("GITHUB_APP_ID")
		}
		if originalPrivateKey != "" {
			os.Setenv("GITHUB_APP_PRIVATE_KEY", originalPrivateKey)
		} else {
			os.Unsetenv("GITHUB_APP_PRIVATE_KEY")
		}
	}()

	// Set test environment variables
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	os.Setenv("GITHUB_APP_ID", "123456")
	os.Setenv("GITHUB_APP_PRIVATE_KEY", "test-key")

	// This should read from environment (but will fail due to fake credentials)
	_, err := NewGitHubAPIWrapper(&Config{})
	// We expect this to fail because we're using fake credentials,
	// but it should get past the initial validation
	if err != nil && strings.Contains(err.Error(), "is required") {
		t.Error("Should have read values from environment variables")
	}
}

func TestIntegrationGetIssues(t *testing.T) {
	// Skip if no real GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping integration test: GitHub credentials not available")
	}

	wrapper, err := NewGitHubAPIWrapper(&Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	result, err := wrapper.GetIssues()
	if err != nil {
		t.Errorf("GetIssues failed: %v", err)
	}

	// Result should be a string (either with issues or "No open issues available")
	if result == "" {
		t.Error("Expected non-empty result from GetIssues")
	}
}

func TestIntegrationListBranches(t *testing.T) {
	// Skip if no real GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping integration test: GitHub credentials not available")
	}

	wrapper, err := NewGitHubAPIWrapper(&Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	result, err := wrapper.ListBranchesInRepo()
	if err != nil {
		t.Errorf("ListBranchesInRepo failed: %v", err)
	}

	// Should have at least one branch (main/master)
	if !strings.Contains(result, "Found") && !strings.Contains(result, "No branches") {
		t.Errorf("Unexpected result format: %s", result)
	}
}

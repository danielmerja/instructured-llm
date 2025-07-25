package agents

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	githubutil "github.com/tmc/langchaingo/util/github"
)

func TestNewGitHubAgentToolkit(t *testing.T) {
	t.Parallel()

	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping test: GitHub credentials not available")
	}

	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	// Test without release tools
	toolkit := NewGitHubAgentToolkit(wrapper)
	require.NotNil(t, toolkit)

	tools := toolkit.GetTools()
	require.Greater(t, len(tools), 10, "Should have multiple tools available")

	// Check that all tools implement the tools.Tool interface
	for _, tool := range tools {
		require.NotEmpty(t, tool.Name(), "Tool should have a name")
		require.NotEmpty(t, tool.Description(), "Tool should have a description")
	}

	// Test with release tools
	toolkitWithReleases := NewGitHubAgentToolkit(wrapper, GitHubAgentToolkitOptions{
		IncludeReleaseTools: true,
	})
	require.NotNil(t, toolkitWithReleases)

	toolsWithReleases := toolkitWithReleases.GetTools()
	require.Greater(t, len(toolsWithReleases), len(tools), "Should have more tools when release tools are included")
}

func TestFromGitHubAPIWrapper(t *testing.T) {
	t.Parallel()

	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping test: GitHub credentials not available")
	}

	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	// Test the factory method
	toolkit := FromGitHubAPIWrapper(wrapper, false)
	require.NotNil(t, toolkit)
	require.False(t, toolkit.includeReleaseTools)

	toolkitWithReleases := FromGitHubAPIWrapper(wrapper, true)
	require.NotNil(t, toolkitWithReleases)
	require.True(t, toolkitWithReleases.includeReleaseTools)
}

func TestGitHubAgentToolkit_GetToolByName(t *testing.T) {
	t.Parallel()

	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping test: GitHub credentials not available")
	}

	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	toolkit := NewGitHubAgentToolkit(wrapper)

	// Test finding existing tool
	tool := toolkit.GetToolByName("Get Issues")
	require.NotNil(t, tool, "Should find 'Get Issues' tool")
	require.Equal(t, "Get Issues", tool.Name())

	// Test non-existent tool
	nonExistentTool := toolkit.GetToolByName("Non-existent Tool")
	require.Nil(t, nonExistentTool, "Should not find non-existent tool")
}

func TestGitHubAgentToolkit_GetToolNames(t *testing.T) {
	t.Parallel()

	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping test: GitHub credentials not available")
	}

	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	toolkit := NewGitHubAgentToolkit(wrapper)
	names := toolkit.GetToolNames()

	require.Greater(t, len(names), 0, "Should have tool names")
	require.Contains(t, names, "Get Issues")
	require.Contains(t, names, "Read File")
	require.Contains(t, names, "Create File")

	// Test with release tools
	toolkitWithReleases := NewGitHubAgentToolkit(wrapper, GitHubAgentToolkitOptions{
		IncludeReleaseTools: true,
	})
	namesWithReleases := toolkitWithReleases.GetToolNames()

	require.Greater(t, len(namesWithReleases), len(names), "Should have more tool names with releases")
	require.Contains(t, namesWithReleases, "Get latest release")
	require.Contains(t, namesWithReleases, "Get releases")
	require.Contains(t, namesWithReleases, "Get release")
}

func TestGitHubAgentTool_Interface(t *testing.T) {
	t.Parallel()

	// Create a mock wrapper for testing
	wrapper := &githubutil.GitHubAPIWrapper{}

	tool := &GitHubAgentTool{
		name:        "Test Tool",
		description: "A test tool for testing",
		wrapper:     wrapper,
		mode:        "test_mode",
	}

	// Test Name method
	require.Equal(t, "Test Tool", tool.Name())

	// Test Description method
	require.Equal(t, "A test tool for testing", tool.Description())

	// Test that it implements tools.Tool interface
	require.Implements(t, (*interface{ Name() string })(nil), tool)
	require.Implements(t, (*interface{ Description() string })(nil), tool)
	require.Implements(t, (*interface {
		Call(context.Context, string) (string, error)
	})(nil), tool)
}

func TestGitHubAgentTool_PreprocessInput(t *testing.T) {
	t.Parallel()

	tool := &GitHubAgentTool{
		name:        "Test Tool",
		description: "A test tool",
		wrapper:     nil, // Not needed for preprocessing tests
		mode:        "get_issue",
	}

	// Test integer extraction for get_issue mode
	result := tool.preprocessInput("42")
	require.Equal(t, "42", result)

	result = tool.preprocessInput("Issue number 123")
	require.Equal(t, "123", result)

	result = tool.preprocessInput("  456  ")
	require.Equal(t, "456", result)

	// Test comment formatting
	tool.mode = "comment_on_issue"
	result = tool.preprocessInput("42\nThis is a comment")
	require.Equal(t, "42\n\nThis is a comment", result)

	// Test PR creation formatting
	tool.mode = "create_pull_request"
	result = tool.preprocessInput("Fix bug\nThis fixes the bug")
	require.Equal(t, "Fix bug\n\nThis fixes the bug", result)

	// Test pass-through for other modes
	tool.mode = "read_file"
	result = tool.preprocessInput("README.md")
	require.Equal(t, "README.md", result)
}

func TestGitHubAgentToolkitIntegration(t *testing.T) {
	t.Parallel()

	// Skip if no GitHub credentials
	if os.Getenv("GITHUB_REPOSITORY") == "" || os.Getenv("GITHUB_APP_PRIVATE_KEY") == "" {
		t.Skip("Skipping integration test: GitHub credentials not available")
	}

	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		t.Skipf("Failed to create wrapper: %v", err)
	}

	toolkit := NewGitHubAgentToolkit(wrapper)
	ctx := context.Background()

	// Test Get Issues tool
	getIssuesTool := toolkit.GetToolByName("Get Issues")
	require.NotNil(t, getIssuesTool)

	result, err := getIssuesTool.Call(ctx, "")
	if err != nil {
		t.Logf("Get Issues call failed (this might be expected): %v", err)
	} else {
		require.NotEmpty(t, result)
		t.Logf("Get Issues result: %s", result)
	}

	// Test Read File tool with a common file
	readFileTool := toolkit.GetToolByName("Read File")
	require.NotNil(t, readFileTool)

	result, err = readFileTool.Call(ctx, "README.md")
	if err != nil {
		t.Logf("Read File call failed (this might be expected if README.md doesn't exist): %v", err)
	} else {
		require.NotEmpty(t, result)
		// Truncate for logging
		if len(result) > 200 {
			result = result[:200] + "..."
		}
		t.Logf("Read File result: %s", result)
	}

	// Test List branches tool
	listBranchesTool := toolkit.GetToolByName("List branches in this repository")
	require.NotNil(t, listBranchesTool)

	result, err = listBranchesTool.Call(ctx, "")
	if err != nil {
		t.Logf("List branches call failed: %v", err)
	} else {
		require.NotEmpty(t, result)
		require.True(t, strings.Contains(result, "branches") || strings.Contains(result, "main") || strings.Contains(result, "master"))
		t.Logf("List branches result: %s", result)
	}
}

func TestGitHubAgentToolkitExpectedTools(t *testing.T) {
	t.Parallel()

	// Create a toolkit with a mock wrapper to test tool creation
	wrapper := &githubutil.GitHubAPIWrapper{}
	toolkit := NewGitHubAgentToolkit(wrapper)

	expectedTools := []string{
		"Get Issues",
		"Get Issue",
		"Comment on Issue",
		"List open pull requests (PRs)",
		"Get Pull Request",
		"Create Pull Request",
		"Create File",
		"Read File",
		"Update File",
		"Delete File",
		"Overview of existing files in Main branch",
		"Overview of files in current working branch",
		"List branches in this repository",
		"Set active branch",
		"Create a new branch",
		"Get files from a directory",
		"Search issues and pull requests",
		"Search code",
	}

	toolNames := toolkit.GetToolNames()

	for _, expectedTool := range expectedTools {
		require.Contains(t, toolNames, expectedTool, "Should contain tool: %s", expectedTool)
	}

	// Test release tools are not included by default
	require.NotContains(t, toolNames, "Get latest release")
	require.NotContains(t, toolNames, "Get releases")
	require.NotContains(t, toolNames, "Get release")

	// Test with release tools
	toolkitWithReleases := NewGitHubAgentToolkit(wrapper, GitHubAgentToolkitOptions{
		IncludeReleaseTools: true,
	})

	releaseToolNames := toolkitWithReleases.GetToolNames()
	require.Contains(t, releaseToolNames, "Get latest release")
	require.Contains(t, releaseToolNames, "Get releases")
	require.Contains(t, releaseToolNames, "Get release")
}

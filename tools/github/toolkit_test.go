package github

import (
	"os"
	"testing"
)

func TestNewToolkit(t *testing.T) {
	// Skip test if required environment variables are not set
	if os.Getenv("GITHUB_TOKEN") == "" || os.Getenv("GITHUB_REPOSITORY") == "" {
		t.Skip("Skipping GitHub toolkit test: GITHUB_TOKEN and GITHUB_REPOSITORY environment variables not set")
	}

	toolkit, err := NewToolkit()
	if err != nil {
		t.Fatalf("Failed to create GitHub toolkit: %v", err)
	}

	tools := toolkit.GetTools()
	if len(tools) == 0 {
		t.Error("Expected toolkit to have tools, but got empty slice")
	}

	// Test that we can get tool names
	names := toolkit.GetToolNames()
	if len(names) != len(tools) {
		t.Errorf("Expected %d tool names, got %d", len(tools), len(names))
	}

	// Test getting a tool by name
	if len(names) > 0 {
		tool := toolkit.GetToolByName(names[0])
		if tool == nil {
			t.Errorf("Expected to find tool with name '%s', but got nil", names[0])
		}
	}

	// Test getting a non-existent tool
	nonExistentTool := toolkit.GetToolByName("NonExistent Tool")
	if nonExistentTool != nil {
		t.Error("Expected nil for non-existent tool, but got a tool")
	}
}

func TestNewToolkitWithReleaseTools(t *testing.T) {
	// Skip test if required environment variables are not set
	if os.Getenv("GITHUB_TOKEN") == "" || os.Getenv("GITHUB_REPOSITORY") == "" {
		t.Skip("Skipping GitHub toolkit test: GITHUB_TOKEN and GITHUB_REPOSITORY environment variables not set")
	}

	toolkit, err := NewToolkit(ToolkitOptions{IncludeReleaseTools: true})
	if err != nil {
		t.Fatalf("Failed to create GitHub toolkit with release tools: %v", err)
	}

	tools := toolkit.GetTools()
	if len(tools) == 0 {
		t.Error("Expected toolkit to have tools, but got empty slice")
	}

	// Check that release tools are included
	names := toolkit.GetToolNames()
	foundReleaseTools := false
	for _, name := range names {
		if name == "Get Releases" || name == "Get Latest Release" || name == "Get Release" {
			foundReleaseTools = true
			break
		}
	}

	if !foundReleaseTools {
		t.Error("Expected to find release tools when IncludeReleaseTools is true")
	}
}

func TestToolkitWithoutEnvironmentVariables(t *testing.T) {
	// Temporarily unset environment variables
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalRepo := os.Getenv("GITHUB_REPOSITORY")

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")

	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		}
		if originalRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalRepo)
		}
	}()

	_, err := NewToolkit()
	if err == nil {
		t.Error("Expected error when creating toolkit without environment variables, but got nil")
	}
}

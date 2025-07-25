package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/tools"
)

// ListBranchesTool lists all branches in the repository.
type ListBranchesTool struct {
	BaseTool
}

var _ tools.Tool = (*ListBranchesTool)(nil)

// NewListBranchesTool creates a new tool for listing branches.
func NewListBranchesTool() (*ListBranchesTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &ListBranchesTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *ListBranchesTool) Name() string {
	return "List Branches"
}

// Description returns the description of the tool.
func (t *ListBranchesTool) Description() string {
	return "This tool will fetch a list of all branches in the repository. It will return the name of each branch. No input parameters are required."
}

// Call executes the tool to list branches.
func (t *ListBranchesTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	branches, _, err := t.client.Repositories.ListBranches(ctx, t.client.Owner(), t.client.Repo(), nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch branches: %w", err)
	}

	var result strings.Builder
	result.WriteString("Repository Branches:\n")
	for _, branch := range branches {
		result.WriteString(fmt.Sprintf("- %s\n", branch.GetName()))
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// GetDirectoryFilesTool lists files in a directory.
type GetDirectoryFilesTool struct {
	BaseTool
}

var _ tools.Tool = (*GetDirectoryFilesTool)(nil)

// NewGetDirectoryFilesTool creates a new tool for listing directory files.
func NewGetDirectoryFilesTool() (*GetDirectoryFilesTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetDirectoryFilesTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetDirectoryFilesTool) Name() string {
	return "Get Directory Files"
}

// Description returns the description of the tool.
func (t *GetDirectoryFilesTool) Description() string {
	return "This tool will fetch a list of all files in a specified directory. **VERY IMPORTANT**: You must specify the path of the directory as a string input parameter."
}

// Call executes the tool to list directory files.
func (t *GetDirectoryFilesTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	dirPath := strings.TrimSpace(input)
	// Remove leading slash if present
	dirPath = strings.TrimPrefix(dirPath, "/")

	_, directoryContent, _, err := t.client.Repositories.GetContents(ctx, t.client.Owner(), t.client.Repo(), dirPath, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch directory contents for %s: %w", dirPath, err)
	}

	var result strings.Builder
	if dirPath == "" {
		result.WriteString("Files in root directory:\n")
	} else {
		result.WriteString(fmt.Sprintf("Files in directory %s:\n", dirPath))
	}

	for _, item := range directoryContent {
		if item.GetType() == "file" {
			result.WriteString(fmt.Sprintf("📄 %s\n", item.GetName()))
		} else if item.GetType() == "dir" {
			result.WriteString(fmt.Sprintf("📁 %s/\n", item.GetName()))
		}
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// SearchCodeTool searches for code in the repository.
type SearchCodeTool struct {
	BaseTool
}

var _ tools.Tool = (*SearchCodeTool)(nil)

// NewSearchCodeTool creates a new tool for searching code.
func NewSearchCodeTool() (*SearchCodeTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &SearchCodeTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *SearchCodeTool) Name() string {
	return "Search Code"
}

// Description returns the description of the tool.
func (t *SearchCodeTool) Description() string {
	return "This tool will search for code in the repository. **VERY IMPORTANT**: You must specify the search query as a string input parameter."
}

// Call executes the tool to search code.
func (t *SearchCodeTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	query := strings.TrimSpace(input)
	if query == "" {
		err := fmt.Errorf("search query cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Add repository qualifier to the search
	searchQuery := fmt.Sprintf("%s repo:%s/%s", query, t.client.Owner(), t.client.Repo())

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	result, _, err := t.client.Search.Code(ctx, searchQuery, opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to search code: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Code search results for '%s':\n\n", query))

	if result.GetTotal() == 0 {
		output.WriteString("No results found.\n")
	} else {
		output.WriteString(fmt.Sprintf("Found %d results:\n\n", result.GetTotal()))
		for _, codeResult := range result.CodeResults {
			output.WriteString(fmt.Sprintf("File: %s\n", codeResult.GetPath()))
			output.WriteString(fmt.Sprintf("Repository: %s\n", codeResult.GetRepository().GetFullName()))
			if codeResult.TextMatches != nil {
				for _, match := range codeResult.TextMatches {
					output.WriteString(fmt.Sprintf("Match: %s\n", match.GetFragment()))
				}
			}
			output.WriteString("\n---\n\n")
		}
	}

	outputStr := output.String()
	t.handleToolEnd(ctx, outputStr)
	return outputStr, nil
}

// SearchIssuesAndPRsTool searches for issues and pull requests.
type SearchIssuesAndPRsTool struct {
	BaseTool
}

var _ tools.Tool = (*SearchIssuesAndPRsTool)(nil)

// NewSearchIssuesAndPRsTool creates a new tool for searching issues and PRs.
func NewSearchIssuesAndPRsTool() (*SearchIssuesAndPRsTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &SearchIssuesAndPRsTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *SearchIssuesAndPRsTool) Name() string {
	return "Search Issues and PRs"
}

// Description returns the description of the tool.
func (t *SearchIssuesAndPRsTool) Description() string {
	return "This tool will search for issues and pull requests in the repository. **VERY IMPORTANT**: You must specify the search query as a string input parameter."
}

// Call executes the tool to search issues and PRs.
func (t *SearchIssuesAndPRsTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	query := strings.TrimSpace(input)
	if query == "" {
		err := fmt.Errorf("search query cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Add repository qualifier to the search
	searchQuery := fmt.Sprintf("%s repo:%s/%s", query, t.client.Owner(), t.client.Repo())

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	result, _, err := t.client.Search.Issues(ctx, searchQuery, opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to search issues and PRs: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s':\n\n", query))

	if result.GetTotal() == 0 {
		output.WriteString("No results found.\n")
	} else {
		output.WriteString(fmt.Sprintf("Found %d results:\n\n", result.GetTotal()))
		for _, issue := range result.Issues {
			if issue.IsPullRequest() {
				output.WriteString(fmt.Sprintf("PR #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
			} else {
				output.WriteString(fmt.Sprintf("Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
			}
			output.WriteString(fmt.Sprintf("State: %s\n", issue.GetState()))
			output.WriteString(fmt.Sprintf("Author: %s\n", issue.GetUser().GetLogin()))
			output.WriteString(fmt.Sprintf("Created: %s\n", issue.GetCreatedAt().Format("2006-01-02 15:04:05")))
			output.WriteString("\n---\n\n")
		}
	}

	outputStr := output.String()
	t.handleToolEnd(ctx, outputStr)
	return outputStr, nil
}

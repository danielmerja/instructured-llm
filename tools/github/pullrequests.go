package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/tools"
)

// ListPullRequestsTool fetches a list of repository pull requests.
type ListPullRequestsTool struct {
	BaseTool
}

var _ tools.Tool = (*ListPullRequestsTool)(nil)

// NewListPullRequestsTool creates a new tool for listing pull requests.
func NewListPullRequestsTool() (*ListPullRequestsTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &ListPullRequestsTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *ListPullRequestsTool) Name() string {
	return "List Pull Requests"
}

// Description returns the description of the tool.
func (t *ListPullRequestsTool) Description() string {
	return "This tool will fetch a list of the repository's Pull Requests (PRs). It will return the title, and PR number of 5 PRs. It takes no input."
}

// Call executes the tool to list pull requests.
func (t *ListPullRequestsTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 5,
		},
	}

	prs, _, err := t.client.PullRequests.List(ctx, t.client.Owner(), t.client.Repo(), opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	var result strings.Builder
	result.WriteString("Repository Pull Requests:\n")
	for _, pr := range prs {
		result.WriteString(fmt.Sprintf("PR #%d: %s\n", pr.GetNumber(), pr.GetTitle()))
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// GetPullRequestTool fetches a specific pull request by number.
type GetPullRequestTool struct {
	BaseTool
}

var _ tools.Tool = (*GetPullRequestTool)(nil)

// NewGetPullRequestTool creates a new tool for getting a specific pull request.
func NewGetPullRequestTool() (*GetPullRequestTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetPullRequestTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetPullRequestTool) Name() string {
	return "Get Pull Request"
}

// Description returns the description of the tool.
func (t *GetPullRequestTool) Description() string {
	return "This tool will fetch the title, body, comment thread and commit history of a specific Pull Request (by PR number). **VERY IMPORTANT**: You must specify the PR number as an integer."
}

// Call executes the tool to get a specific pull request.
func (t *GetPullRequestTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	prNumber, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("invalid PR number: %s", input)
	}

	pr, _, err := t.client.PullRequests.Get(ctx, t.client.Owner(), t.client.Repo(), prNumber)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch PR #%d: %w", prNumber, err)
	}

	// Get comments
	comments, _, err := t.client.Issues.ListComments(ctx, t.client.Owner(), t.client.Repo(), prNumber, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch comments for PR #%d: %w", prNumber, err)
	}

	// Get commits
	commits, _, err := t.client.PullRequests.ListCommits(ctx, t.client.Owner(), t.client.Repo(), prNumber, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch commits for PR #%d: %w", prNumber, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pull Request #%d: %s\n\n", pr.GetNumber(), pr.GetTitle()))
	result.WriteString(fmt.Sprintf("State: %s\n", pr.GetState()))
	result.WriteString(fmt.Sprintf("Author: %s\n", pr.GetUser().GetLogin()))
	result.WriteString(fmt.Sprintf("Created: %s\n", pr.GetCreatedAt().Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Base: %s <- Head: %s\n\n", pr.GetBase().GetRef(), pr.GetHead().GetRef()))

	body := pr.GetBody()
	if body != "" {
		result.WriteString("Body:\n")
		result.WriteString(body)
		result.WriteString("\n\n")
	}

	if len(commits) > 0 {
		result.WriteString("Commits:\n")
		for _, commit := range commits {
			result.WriteString(fmt.Sprintf("- %s: %s\n",
				commit.GetSHA()[:8],
				commit.GetCommit().GetMessage()))
		}
		result.WriteString("\n")
	}

	if len(comments) > 0 {
		result.WriteString("Comments:\n")
		for i, comment := range comments {
			result.WriteString(fmt.Sprintf("Comment #%d by %s (%s):\n", i+1,
				comment.GetUser().GetLogin(),
				comment.GetCreatedAt().Format("2006-01-02 15:04:05")))
			result.WriteString(comment.GetBody())
			result.WriteString("\n\n")
		}
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// CreatePullRequestTool creates a new pull request.
type CreatePullRequestTool struct {
	BaseTool
}

var _ tools.Tool = (*CreatePullRequestTool)(nil)

// NewCreatePullRequestTool creates a new tool for creating pull requests.
func NewCreatePullRequestTool() (*CreatePullRequestTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &CreatePullRequestTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *CreatePullRequestTool) Name() string {
	return "Create Pull Request"
}

// Description returns the description of the tool.
func (t *CreatePullRequestTool) Description() string {
	return `This tool is useful when you need to create a new pull request in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules:

- First you must specify the title of the pull request
- Then you must place two newlines
- Then you must write the body or description of the pull request

When appropriate, always reference relevant issues in the body by using the syntax ` + "`closes #<issue_number>`" + ` like ` + "`closes #3, closes #6`" + `.
For example, if you would like to create a pull request called "README updates" with contents "added contributors' names, closes #3", you would pass in the following string:

README updates

added contributors' names, closes #3`
}

// Call executes the tool to create a pull request.
func (t *CreatePullRequestTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	parts := strings.SplitN(input, "\n\n", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("invalid input format: expected 'title\\n\\nbody', got: %s", input)
		t.handleToolError(ctx, err)
		return "", err
	}

	title := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])

	if title == "" {
		err := fmt.Errorf("pull request title cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Get the default branch to use as base
	repo, _, err := t.client.Repositories.Get(ctx, t.client.Owner(), t.client.Repo())
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to get repository info: %w", err)
	}

	// For simplicity, assume we're creating a PR from the current HEAD to the default branch
	// In a real scenario, you might want to get the current branch name
	head := "HEAD"
	base := repo.GetDefaultBranch()

	newPR := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
	}

	pr, _, err := t.client.PullRequests.Create(ctx, t.client.Owner(), t.client.Repo(), newPR)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	result := fmt.Sprintf("Successfully created pull request #%d: %s",
		pr.GetNumber(), pr.GetTitle())

	t.handleToolEnd(ctx, result)
	return result, nil
}

// ListPullRequestFilesTool lists files in a pull request.
type ListPullRequestFilesTool struct {
	BaseTool
}

var _ tools.Tool = (*ListPullRequestFilesTool)(nil)

// NewListPullRequestFilesTool creates a new tool for listing PR files.
func NewListPullRequestFilesTool() (*ListPullRequestFilesTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &ListPullRequestFilesTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *ListPullRequestFilesTool) Name() string {
	return "List Pull Request Files"
}

// Description returns the description of the tool.
func (t *ListPullRequestFilesTool) Description() string {
	return "This tool will fetch the full text of all files in a pull request (PR) given the PR number as an input. This is useful for understanding the code changes in a PR or contributing to it. **VERY IMPORTANT**: You must specify the PR number as an integer input parameter."
}

// Call executes the tool to list pull request files.
func (t *ListPullRequestFilesTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	prNumber, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("invalid PR number: %s", input)
	}

	files, _, err := t.client.PullRequests.ListFiles(ctx, t.client.Owner(), t.client.Repo(), prNumber, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch files for PR #%d: %w", prNumber, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files in Pull Request #%d:\n\n", prNumber))

	for _, file := range files {
		result.WriteString(fmt.Sprintf("File: %s\n", file.GetFilename()))
		result.WriteString(fmt.Sprintf("Status: %s\n", file.GetStatus()))
		result.WriteString(fmt.Sprintf("Additions: %d, Deletions: %d, Changes: %d\n",
			file.GetAdditions(), file.GetDeletions(), file.GetChanges()))

		if file.GetPatch() != "" {
			result.WriteString("Patch:\n")
			result.WriteString(file.GetPatch())
		}
		result.WriteString("\n---\n\n")
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

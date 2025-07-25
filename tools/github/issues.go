package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/tools"
)

// GetIssuesTool fetches a list of repository issues.
type GetIssuesTool struct {
	BaseTool
}

var _ tools.Tool = (*GetIssuesTool)(nil)

// NewGetIssuesTool creates a new tool for getting repository issues.
func NewGetIssuesTool() (*GetIssuesTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetIssuesTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetIssuesTool) Name() string {
	return "Get Issues"
}

// Description returns the description of the tool.
func (t *GetIssuesTool) Description() string {
	return "This tool will fetch a list of the repository's issues. It will return the title, and issue number of 5 issues. It takes no input."
}

// Call executes the tool to get repository issues.
func (t *GetIssuesTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	opts := &github.IssueListByRepoOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 5,
		},
	}

	issues, _, err := t.client.Issues.ListByRepo(ctx, t.client.Owner(), t.client.Repo(), opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch issues: %w", err)
	}

	var result strings.Builder
	result.WriteString("Repository Issues:\n")
	for _, issue := range issues {
		result.WriteString(fmt.Sprintf("Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// GetIssueTool fetches a specific issue by number.
type GetIssueTool struct {
	BaseTool
}

var _ tools.Tool = (*GetIssueTool)(nil)

// NewGetIssueTool creates a new tool for getting a specific issue.
func NewGetIssueTool() (*GetIssueTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetIssueTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetIssueTool) Name() string {
	return "Get Issue"
}

// Description returns the description of the tool.
func (t *GetIssueTool) Description() string {
	return "This tool will fetch the title, body, and comment thread of a specific issue. **VERY IMPORTANT**: You must specify the issue number as an integer."
}

// Call executes the tool to get a specific issue.
func (t *GetIssueTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	issueNumber, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("invalid issue number: %s", input)
	}

	issue, _, err := t.client.Issues.Get(ctx, t.client.Owner(), t.client.Repo(), issueNumber)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch issue #%d: %w", issueNumber, err)
	}

	// Get comments
	comments, _, err := t.client.Issues.ListComments(ctx, t.client.Owner(), t.client.Repo(), issueNumber, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch comments for issue #%d: %w", issueNumber, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Issue #%d: %s\n\n", issue.GetNumber(), issue.GetTitle()))
	result.WriteString(fmt.Sprintf("State: %s\n", issue.GetState()))
	result.WriteString(fmt.Sprintf("Author: %s\n", issue.GetUser().GetLogin()))
	result.WriteString(fmt.Sprintf("Created: %s\n\n", issue.GetCreatedAt().Format("2006-01-02 15:04:05")))

	body := issue.GetBody()
	if body != "" {
		result.WriteString("Body:\n")
		result.WriteString(body)
		result.WriteString("\n\n")
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

// CommentOnIssueTool creates a comment on a specific issue.
type CommentOnIssueTool struct {
	BaseTool
}

var _ tools.Tool = (*CommentOnIssueTool)(nil)

// NewCommentOnIssueTool creates a new tool for commenting on issues.
func NewCommentOnIssueTool() (*CommentOnIssueTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &CommentOnIssueTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *CommentOnIssueTool) Name() string {
	return "Comment on Issue"
}

// Description returns the description of the tool.
func (t *CommentOnIssueTool) Description() string {
	return `This tool is useful when you need to comment on a GitHub issue. Simply pass in the issue number and the comment you would like to make. Please use this sparingly as we don't want to clutter the comment threads. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules:

- First you must specify the issue number as an integer
- Then you must place two newlines
- Then you must specify your comment`
}

// Call executes the tool to comment on an issue.
func (t *CommentOnIssueTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	parts := strings.SplitN(input, "\n\n", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("invalid input format: expected 'issue_number\\n\\ncomment', got: %s", input)
		t.handleToolError(ctx, err)
		return "", err
	}

	issueNumber, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("invalid issue number: %s", parts[0])
	}

	commentBody := strings.TrimSpace(parts[1])
	if commentBody == "" {
		err := fmt.Errorf("comment body cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	comment := &github.IssueComment{
		Body: &commentBody,
	}

	createdComment, _, err := t.client.Issues.CreateComment(ctx, t.client.Owner(), t.client.Repo(), issueNumber, comment)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to create comment on issue #%d: %w", issueNumber, err)
	}

	result := fmt.Sprintf("Successfully created comment #%d on issue #%d",
		createdComment.GetID(), issueNumber)

	t.handleToolEnd(ctx, result)
	return result, nil
}

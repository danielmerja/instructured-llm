package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/callbacks"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub client with additional configuration.
type Client struct {
	*github.Client
	owner string
	repo  string
}

// NewClient creates a new GitHub client from environment variables.
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable is required")
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return nil, errors.New("GITHUB_REPOSITORY environment variable is required (format: owner/repo)")
	}

	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("GITHUB_REPOSITORY must be in format 'owner/repo', got: %s", repository)
	}

	owner, repo := parts[0], parts[1]

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	return &Client{
		Client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

// Owner returns the repository owner.
func (c *Client) Owner() string {
	return c.owner
}

// Repo returns the repository name.
func (c *Client) Repo() string {
	return c.repo
}

// BaseTool provides a base implementation for GitHub tools.
type BaseTool struct {
	client           *Client
	callbacksHandler callbacks.Handler
}

// SetCallbacksHandler sets the callbacks handler for the tool.
func (bt *BaseTool) SetCallbacksHandler(handler callbacks.Handler) {
	bt.callbacksHandler = handler
}

// handleToolStart calls the tool start callback if a handler is set.
func (bt *BaseTool) handleToolStart(ctx context.Context, input string) {
	if bt.callbacksHandler != nil {
		bt.callbacksHandler.HandleToolStart(ctx, input)
	}
}

// handleToolEnd calls the tool end callback if a handler is set.
func (bt *BaseTool) handleToolEnd(ctx context.Context, output string) {
	if bt.callbacksHandler != nil {
		bt.callbacksHandler.HandleToolEnd(ctx, output)
	}
}

// handleToolError calls the tool error callback if a handler is set.
func (bt *BaseTool) handleToolError(ctx context.Context, err error) {
	if bt.callbacksHandler != nil {
		bt.callbacksHandler.HandleToolError(ctx, err)
	}
}

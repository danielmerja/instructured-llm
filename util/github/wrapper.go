package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	githubapi "github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// GitHubAPIWrapper wraps the GitHub API with high-level operations.
type GitHubAPIWrapper struct {
	client           *githubapi.Client
	repo             *githubapi.Repository
	owner            string
	repoName         string
	activeBranch     string
	githubBaseBranch string
	appID            string
	privateKey       string
}

// Config holds configuration for the GitHub API wrapper.
type Config struct {
	Repository       string // Format: "owner/repo"
	AppID            string // GitHub App ID
	PrivateKey       string // GitHub App private key (content or file path)
	ActiveBranch     string // Current working branch
	GitHubBaseBranch string // Base branch (usually main/master)
}

// NewGitHubAPIWrapper creates a new GitHub API wrapper with App authentication.
func NewGitHubAPIWrapper(config *Config) (*GitHubAPIWrapper, error) {
	if config == nil {
		config = &Config{}
	}

	// Get values from environment if not provided
	if config.Repository == "" {
		config.Repository = os.Getenv("GITHUB_REPOSITORY")
	}
	if config.AppID == "" {
		config.AppID = os.Getenv("GITHUB_APP_ID")
	}
	if config.PrivateKey == "" {
		config.PrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
	}

	if config.Repository == "" {
		return nil, errors.New("GITHUB_REPOSITORY is required")
	}
	if config.AppID == "" {
		return nil, errors.New("GITHUB_APP_ID is required")
	}
	if config.PrivateKey == "" {
		return nil, errors.New("GITHUB_APP_PRIVATE_KEY is required")
	}

	// Parse repository
	parts := strings.Split(config.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", config.Repository)
	}
	owner, repoName := parts[0], parts[1]

	// Set up GitHub App authentication
	// Note: This is a simplified version. For production, you'd want to implement
	// proper GitHub App authentication with JWT tokens and installation tokens
	client := githubapi.NewClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: config.PrivateKey, // Assuming token for now
			}),
		},
	})

	// Get repository info
	repo, _, err := client.Repositories.Get(context.Background(), owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Set default branches
	if config.GitHubBaseBranch == "" {
		config.GitHubBaseBranch = repo.GetDefaultBranch()
	}
	if config.ActiveBranch == "" {
		config.ActiveBranch = repo.GetDefaultBranch()
	}

	return &GitHubAPIWrapper{
		client:           client,
		repo:             repo,
		owner:            owner,
		repoName:         repoName,
		activeBranch:     config.ActiveBranch,
		githubBaseBranch: config.GitHubBaseBranch,
		appID:            config.AppID,
		privateKey:       config.PrivateKey,
	}, nil
}

// Issue represents a parsed GitHub issue.
type Issue struct {
	Title    string `json:"title"`
	Number   int    `json:"number"`
	OpenedBy string `json:"opened_by,omitempty"`
}

// PullRequest represents a parsed GitHub pull request.
type PullRequest struct {
	Title    string `json:"title"`
	Number   int    `json:"number"`
	Commits  string `json:"commits"`
	Comments string `json:"comments"`
}

// ParseIssues extracts title, number, and opener from GitHub issues.
func (w *GitHubAPIWrapper) ParseIssues(issues []*githubapi.Issue) []Issue {
	var parsed []Issue
	for _, issue := range issues {
		i := Issue{
			Title:  issue.GetTitle(),
			Number: issue.GetNumber(),
		}
		if issue.GetUser() != nil {
			i.OpenedBy = issue.GetUser().GetLogin()
		}
		parsed = append(parsed, i)
	}
	return parsed
}

// ParsePullRequests extracts relevant data from GitHub pull requests.
func (w *GitHubAPIWrapper) ParsePullRequests(prs []*githubapi.PullRequest) []PullRequest {
	var parsed []PullRequest
	for _, pr := range prs {
		parsed = append(parsed, PullRequest{
			Title:    pr.GetTitle(),
			Number:   pr.GetNumber(),
			Commits:  strconv.Itoa(pr.GetCommits()),
			Comments: strconv.Itoa(pr.GetComments()),
		})
	}
	return parsed
}

// GetIssues fetches all open issues from the repository excluding pull requests.
func (w *GitHubAPIWrapper) GetIssues() (string, error) {
	opts := &githubapi.IssueListByRepoOptions{
		State: "open",
	}

	issues, _, err := w.client.Issues.ListByRepo(context.Background(), w.owner, w.repoName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to fetch issues: %w", err)
	}

	// Filter out pull requests
	var filteredIssues []*githubapi.Issue
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	if len(filteredIssues) == 0 {
		return "No open issues available", nil
	}

	parsed := w.ParseIssues(filteredIssues)
	return fmt.Sprintf("Found %d issues:\n%+v", len(parsed), parsed), nil
}

// ListOpenPullRequests fetches all open PRs from the repository.
func (w *GitHubAPIWrapper) ListOpenPullRequests() (string, error) {
	opts := &githubapi.PullRequestListOptions{
		State: "open",
	}

	prs, _, err := w.client.PullRequests.List(context.Background(), w.owner, w.repoName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	if len(prs) == 0 {
		return "No open pull requests available", nil
	}

	parsed := w.ParsePullRequests(prs)
	return fmt.Sprintf("Found %d pull requests:\n%+v", len(parsed), parsed), nil
}

// ListFilesInMainBranch fetches all files in the main branch of the repository.
func (w *GitHubAPIWrapper) ListFilesInMainBranch() (string, error) {
	files, err := w.listFiles("", w.githubBaseBranch)
	if err != nil {
		return "", fmt.Errorf("failed to list files in main branch: %w", err)
	}

	if len(files) == 0 {
		return "No files found in the main branch", nil
	}

	return fmt.Sprintf("Found %d files in the main branch:\n%s", len(files), strings.Join(files, "\n")), nil
}

// SetActiveBranch sets the active branch for the wrapper.
func (w *GitHubAPIWrapper) SetActiveBranch(branchName string) (string, error) {
	branches, _, err := w.client.Repositories.ListBranches(context.Background(), w.owner, w.repoName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	var branchNames []string
	branchExists := false
	for _, branch := range branches {
		name := branch.GetName()
		branchNames = append(branchNames, name)
		if name == branchName {
			branchExists = true
		}
	}

	if !branchExists {
		return fmt.Sprintf("Error %s does not exist, in repo with current branches: %v", branchName, branchNames), nil
	}

	w.activeBranch = branchName
	return fmt.Sprintf("Switched to branch `%s`", branchName), nil
}

// ListBranchesInRepo fetches a list of all branches in the repository.
func (w *GitHubAPIWrapper) ListBranchesInRepo() (string, error) {
	branches, _, err := w.client.Repositories.ListBranches(context.Background(), w.owner, w.repoName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	if len(branches) == 0 {
		return "No branches found in the repository", nil
	}

	var branchNames []string
	for _, branch := range branches {
		branchNames = append(branchNames, branch.GetName())
	}

	return fmt.Sprintf("Found %d branches in the repository:\n%s", len(branchNames), strings.Join(branchNames, "\n")), nil
}

// CreateBranch creates a new branch and sets it as the active branch.
func (w *GitHubAPIWrapper) CreateBranch(proposedBranchName string) (string, error) {
	// Get the base branch reference
	baseBranch, _, err := w.client.Git.GetRef(context.Background(), w.owner, w.repoName, "refs/heads/"+w.githubBaseBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get base branch: %w", err)
	}

	newBranchName := proposedBranchName
	for i := 0; i < 1000; i++ {
		ref := &githubapi.Reference{
			Ref: githubapi.String("refs/heads/" + newBranchName),
			Object: &githubapi.GitObject{
				SHA: baseBranch.Object.SHA,
			},
		}

		_, _, err := w.client.Git.CreateRef(context.Background(), w.owner, w.repoName, ref)
		if err == nil {
			w.activeBranch = newBranchName
			return fmt.Sprintf("Branch '%s' created successfully, and set as current active branch.", newBranchName), nil
		}

		// If branch already exists, try with a version suffix
		if strings.Contains(err.Error(), "Reference already exists") {
			newBranchName = fmt.Sprintf("%s_v%d", proposedBranchName, i+1)
			continue
		}

		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	return fmt.Sprintf("Unable to create branch. At least 1000 branches exist with named derived from proposed_branch_name: `%s`", proposedBranchName), nil
}

// ListFilesInBotBranch fetches all files in the active branch.
func (w *GitHubAPIWrapper) ListFilesInBotBranch() (string, error) {
	files, err := w.listFiles("", w.activeBranch)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if len(files) == 0 {
		return fmt.Sprintf("No files found in branch: `%s`", w.activeBranch), nil
	}

	return fmt.Sprintf("Found %d files in branch `%s`:\n%s", len(files), w.activeBranch, strings.Join(files, "\n")), nil
}

// GetFilesFromDirectory recursively fetches files from a directory.
func (w *GitHubAPIWrapper) GetFilesFromDirectory(directoryPath string) (string, error) {
	files, err := w.listFiles(directoryPath, w.activeBranch)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return strings.Join(files, "\n"), nil
}

// listFiles is a helper function to recursively list files.
func (w *GitHubAPIWrapper) listFiles(path, branch string) ([]string, error) {
	var files []string

	_, contents, _, err := w.client.Repositories.GetContents(context.Background(), w.owner, w.repoName, path, &githubapi.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		return nil, err
	}

	for _, content := range contents {
		if content.GetType() == "dir" {
			subFiles, err := w.listFiles(content.GetPath(), branch)
			if err != nil {
				continue // Skip directories that can't be read
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, content.GetPath())
		}
	}

	return files, nil
}

// GetIssue fetches a specific issue and its first 10 comments.
func (w *GitHubAPIWrapper) GetIssue(issueNumber int) (map[string]interface{}, error) {
	issue, _, err := w.client.Issues.Get(context.Background(), w.owner, w.repoName, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Get comments (first 10)
	comments, _, err := w.client.Issues.ListComments(context.Background(), w.owner, w.repoName, issueNumber, &githubapi.IssueListCommentsOptions{
		ListOptions: githubapi.ListOptions{PerPage: 10},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	var commentsList []map[string]string
	for _, comment := range comments {
		commentsList = append(commentsList, map[string]string{
			"body": comment.GetBody(),
			"user": comment.GetUser().GetLogin(),
		})
	}

	openedBy := ""
	if issue.GetUser() != nil {
		openedBy = issue.GetUser().GetLogin()
	}

	commentsJSON, _ := json.Marshal(commentsList)

	return map[string]interface{}{
		"number":    issueNumber,
		"title":     issue.GetTitle(),
		"body":      issue.GetBody(),
		"comments":  string(commentsJSON),
		"opened_by": openedBy,
	}, nil
}

// GetPullRequest fetches a specific pull request with comments and commits.
func (w *GitHubAPIWrapper) GetPullRequest(prNumber int) (map[string]interface{}, error) {
	pr, _, err := w.client.PullRequests.Get(context.Background(), w.owner, w.repoName, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	result := map[string]interface{}{
		"title":  pr.GetTitle(),
		"number": strconv.Itoa(prNumber),
		"body":   pr.GetBody(),
	}

	// Get comments (first 10)
	comments, _, err := w.client.Issues.ListComments(context.Background(), w.owner, w.repoName, prNumber, &githubapi.IssueListCommentsOptions{
		ListOptions: githubapi.ListOptions{PerPage: 10},
	})
	if err == nil {
		var commentsList []string
		for _, comment := range comments {
			commentData := map[string]string{
				"body": comment.GetBody(),
				"user": comment.GetUser().GetLogin(),
			}
			commentJSON, _ := json.Marshal(commentData)
			commentsList = append(commentsList, string(commentJSON))
		}
		commentsJSON, _ := json.Marshal(commentsList)
		result["comments"] = string(commentsJSON)
	}

	// Get commits (first 10)
	commits, _, err := w.client.PullRequests.ListCommits(context.Background(), w.owner, w.repoName, prNumber, &githubapi.ListOptions{PerPage: 10})
	if err == nil {
		var commitsList []string
		for _, commit := range commits {
			commitData := map[string]string{
				"message": commit.GetCommit().GetMessage(),
			}
			commitJSON, _ := json.Marshal(commitData)
			commitsList = append(commitsList, string(commitJSON))
		}
		commitsJSON, _ := json.Marshal(commitsList)
		result["commits"] = string(commitsJSON)
	}

	return result, nil
}

// CreatePullRequest creates a pull request from the active branch to the base branch.
func (w *GitHubAPIWrapper) CreatePullRequest(prQuery string) (string, error) {
	if w.githubBaseBranch == w.activeBranch {
		return "Cannot make a pull request because commits are already in the main or master branch.", nil
	}

	lines := strings.Split(prQuery, "\n")
	if len(lines) < 1 {
		return "Invalid PR query format", nil
	}

	title := lines[0]
	body := ""
	if len(lines) > 2 {
		body = strings.Join(lines[2:], "\n")
	}

	newPR := &githubapi.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &w.activeBranch,
		Base:  &w.githubBaseBranch,
	}

	pr, _, err := w.client.PullRequests.Create(context.Background(), w.owner, w.repoName, newPR)
	if err != nil {
		return fmt.Sprintf("Unable to make pull request due to error:\n%v", err), nil
	}

	return fmt.Sprintf("Successfully created PR number %d", pr.GetNumber()), nil
}

// CommentOnIssue adds a comment to a GitHub issue.
func (w *GitHubAPIWrapper) CommentOnIssue(commentQuery string) (string, error) {
	parts := strings.SplitN(commentQuery, "\n\n", 2)
	if len(parts) != 2 {
		return "Invalid comment format", nil
	}

	issueNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Sprintf("Invalid issue number: %s", parts[0]), nil
	}

	comment := parts[1]

	issueComment := &githubapi.IssueComment{
		Body: &comment,
	}

	_, _, err = w.client.Issues.CreateComment(context.Background(), w.owner, w.repoName, issueNumber, issueComment)
	if err != nil {
		return fmt.Sprintf("Unable to make comment due to error:\n%v", err), nil
	}

	return fmt.Sprintf("Commented on issue %d", issueNumber), nil
}

// ReadFile reads a file from the active branch.
func (w *GitHubAPIWrapper) ReadFile(filePath string) (string, error) {
	fileContent, _, _, err := w.client.Repositories.GetContents(context.Background(), w.owner, w.repoName, filePath, &githubapi.RepositoryContentGetOptions{
		Ref: w.activeBranch,
	})
	if err != nil {
		return fmt.Sprintf("File not found `%s` on branch `%s`. Error: %v", filePath, w.activeBranch, err), nil
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Sprintf("Failed to decode file content: %v", err), nil
	}

	return content, nil
}

// CreateFile creates a new file in the repository.
func (w *GitHubAPIWrapper) CreateFile(fileQuery string) (string, error) {
	if w.activeBranch == w.githubBaseBranch {
		return fmt.Sprintf("You're attempting to commit to the directly to the %s branch, which is protected. Please create a new branch and try again.", w.githubBaseBranch), nil
	}

	lines := strings.SplitN(fileQuery, "\n", 2)
	if len(lines) < 2 {
		return "Invalid file format", nil
	}

	filePath := lines[0]
	fileContents := lines[1]

	// Check if file already exists
	_, _, _, err := w.client.Repositories.GetContents(context.Background(), w.owner, w.repoName, filePath, &githubapi.RepositoryContentGetOptions{
		Ref: w.activeBranch,
	})
	if err == nil {
		return fmt.Sprintf("File already exists at `%s` on branch `%s`. You must use `update_file` to modify it.", filePath, w.activeBranch), nil
	}

	message := "Create " + filePath
	opts := &githubapi.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(fileContents),
		Branch:  &w.activeBranch,
	}

	_, _, err = w.client.Repositories.CreateFile(context.Background(), w.owner, w.repoName, filePath, opts)
	if err != nil {
		return fmt.Sprintf("Unable to make file due to error:\n%v", err), nil
	}

	return fmt.Sprintf("Created file %s", filePath), nil
}

// UpdateFile updates a file with new content.
func (w *GitHubAPIWrapper) UpdateFile(fileQuery string) (string, error) {
	if w.activeBranch == w.githubBaseBranch {
		return fmt.Sprintf("You're attempting to commit to the directly to the %s branch, which is protected. Please create a new branch and try again.", w.githubBaseBranch), nil
	}

	lines := strings.Split(fileQuery, "\n")
	if len(lines) < 1 {
		return "Invalid file format", nil
	}

	filePath := lines[0]
	content := strings.Join(lines[1:], "\n")

	// Extract old and new content
	oldStartIdx := strings.Index(content, "OLD <<<<")
	oldEndIdx := strings.Index(content, ">>>> OLD")
	newStartIdx := strings.Index(content, "NEW <<<<")
	newEndIdx := strings.Index(content, ">>>> NEW")

	if oldStartIdx == -1 || oldEndIdx == -1 || newStartIdx == -1 || newEndIdx == -1 {
		return "Invalid update format: missing OLD <<<< ... >>>> OLD or NEW <<<< ... >>>> NEW markers", nil
	}

	oldContent := strings.TrimSpace(content[oldStartIdx+8 : oldEndIdx])
	newContent := strings.TrimSpace(content[newStartIdx+8 : newEndIdx])

	// Get current file content
	currentContent, err := w.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Failed to read current file: %v", err), nil
	}

	// Replace content
	updatedContent := strings.ReplaceAll(currentContent, oldContent, newContent)
	if currentContent == updatedContent {
		return "File content was not updated because old content was not found. It may be helpful to use the read_file action to get the current file contents.", nil
	}

	// Get file SHA for update
	fileContent, _, _, err := w.client.Repositories.GetContents(context.Background(), w.owner, w.repoName, filePath, &githubapi.RepositoryContentGetOptions{
		Ref: w.activeBranch,
	})
	if err != nil {
		return fmt.Sprintf("Failed to get file SHA: %v", err), nil
	}

	message := "Update " + filePath
	opts := &githubapi.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(updatedContent),
		Branch:  &w.activeBranch,
		SHA:     fileContent.SHA,
	}

	_, _, err = w.client.Repositories.UpdateFile(context.Background(), w.owner, w.repoName, filePath, opts)
	if err != nil {
		return fmt.Sprintf("Unable to update file due to error:\n%v", err), nil
	}

	return fmt.Sprintf("Updated file %s", filePath), nil
}

// DeleteFile deletes a file from the repository.
func (w *GitHubAPIWrapper) DeleteFile(filePath string) (string, error) {
	if w.activeBranch == w.githubBaseBranch {
		return fmt.Sprintf("You're attempting to commit to the directly to the %s branch, which is protected. Please create a new branch and try again.", w.githubBaseBranch), nil
	}

	// Get file SHA for deletion
	fileContent, _, _, err := w.client.Repositories.GetContents(context.Background(), w.owner, w.repoName, filePath, &githubapi.RepositoryContentGetOptions{
		Ref: w.activeBranch,
	})
	if err != nil {
		return fmt.Sprintf("Unable to delete file due to error:\n%v", err), nil
	}

	message := "Delete " + filePath
	opts := &githubapi.RepositoryContentFileOptions{
		Message: &message,
		Branch:  &w.activeBranch,
		SHA:     fileContent.SHA,
	}

	_, _, err = w.client.Repositories.DeleteFile(context.Background(), w.owner, w.repoName, filePath, opts)
	if err != nil {
		return fmt.Sprintf("Unable to delete file due to error:\n%v", err), nil
	}

	return fmt.Sprintf("Deleted file %s", filePath), nil
}

// SearchIssuesAndPRs searches issues and pull requests in the repository.
func (w *GitHubAPIWrapper) SearchIssuesAndPRs(query string) (string, error) {
	searchQuery := fmt.Sprintf("%s repo:%s/%s", query, w.owner, w.repoName)

	opts := &githubapi.SearchOptions{
		ListOptions: githubapi.ListOptions{PerPage: 5},
	}

	result, _, err := w.client.Search.Issues(context.Background(), searchQuery, opts)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err), nil
	}

	if result.GetTotal() == 0 {
		return "0 results found.", nil
	}

	maxResults := 5
	if result.GetTotal() < 5 {
		maxResults = result.GetTotal()
	}

	var results []string
	results = append(results, fmt.Sprintf("Top %d results:", maxResults))

	for i, issue := range result.Issues {
		if i >= maxResults {
			break
		}
		results = append(results, fmt.Sprintf("Title: %s, Number: %d, State: %s", issue.GetTitle(), issue.GetNumber(), issue.GetState()))
	}

	return strings.Join(results, "\n"), nil
}

// SearchCode searches code in the repository.
func (w *GitHubAPIWrapper) SearchCode(query string) (string, error) {
	searchQuery := fmt.Sprintf("%s repo:%s/%s", query, w.owner, w.repoName)

	opts := &githubapi.SearchOptions{
		ListOptions: githubapi.ListOptions{PerPage: 5},
	}

	result, _, err := w.client.Search.Code(context.Background(), searchQuery, opts)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err), nil
	}

	if result.GetTotal() == 0 {
		return "0 results found.", nil
	}

	maxResults := 5
	if result.GetTotal() < 5 {
		maxResults = result.GetTotal()
	}

	var results []string
	results = append(results, fmt.Sprintf("Showing top %d of %d results:", maxResults, result.GetTotal()))

	for i, code := range result.CodeResults {
		if i >= maxResults {
			break
		}

		// Get file content
		content, err := w.ReadFile(code.GetPath())
		if err != nil {
			content = fmt.Sprintf("Error reading file: %v", err)
		}

		results = append(results, fmt.Sprintf("Filepath: `%s`\nFile contents: %s\n<END OF FILE>", code.GetPath(), content))
	}

	return strings.Join(results, "\n"), nil
}

// GetLatestRelease fetches the latest release of the repository.
func (w *GitHubAPIWrapper) GetLatestRelease() (string, error) {
	release, _, err := w.client.Repositories.GetLatestRelease(context.Background(), w.owner, w.repoName)
	if err != nil {
		return fmt.Sprintf("Failed to get latest release: %v", err), nil
	}

	return fmt.Sprintf("Latest title: %s tag: %s body: %s", release.GetName(), release.GetTagName(), release.GetBody()), nil
}

// GetReleases fetches releases of the repository.
func (w *GitHubAPIWrapper) GetReleases() (string, error) {
	opts := &githubapi.ListOptions{PerPage: 5}
	releases, _, err := w.client.Repositories.ListReleases(context.Background(), w.owner, w.repoName, opts)
	if err != nil {
		return fmt.Sprintf("Failed to get releases: %v", err), nil
	}

	if len(releases) == 0 {
		return "No releases found.", nil
	}

	var results []string
	results = append(results, fmt.Sprintf("Top %d results:", len(releases)))

	for _, release := range releases {
		results = append(results, fmt.Sprintf("Title: %s, Tag: %s, Body: %s", release.GetName(), release.GetTagName(), release.GetBody()))
	}

	return strings.Join(results, "\n"), nil
}

// GetRelease fetches a specific release by tag name.
func (w *GitHubAPIWrapper) GetRelease(tagName string) (string, error) {
	release, _, err := w.client.Repositories.GetReleaseByTag(context.Background(), w.owner, w.repoName, tagName)
	if err != nil {
		return fmt.Sprintf("Failed to get release: %v", err), nil
	}

	return fmt.Sprintf("Release: %s tag: %s body: %s", release.GetName(), release.GetTagName(), release.GetBody()), nil
}

// Run executes a GitHub operation based on the mode and query.
func (w *GitHubAPIWrapper) Run(mode, query string) (string, error) {
	switch mode {
	case "get_issue":
		issueNum, err := strconv.Atoi(query)
		if err != nil {
			return "", fmt.Errorf("invalid issue number: %s", query)
		}
		result, err := w.GetIssue(issueNum)
		if err != nil {
			return "", err
		}
		jsonData, _ := json.Marshal(result)
		return string(jsonData), nil

	case "get_pull_request":
		prNum, err := strconv.Atoi(query)
		if err != nil {
			return "", fmt.Errorf("invalid PR number: %s", query)
		}
		result, err := w.GetPullRequest(prNum)
		if err != nil {
			return "", err
		}
		jsonData, _ := json.Marshal(result)
		return string(jsonData), nil

	case "get_issues":
		return w.GetIssues()

	case "comment_on_issue":
		return w.CommentOnIssue(query)

	case "create_file":
		return w.CreateFile(query)

	case "create_pull_request":
		return w.CreatePullRequest(query)

	case "read_file":
		return w.ReadFile(query)

	case "update_file":
		return w.UpdateFile(query)

	case "delete_file":
		return w.DeleteFile(query)

	case "list_open_pull_requests":
		return w.ListOpenPullRequests()

	case "list_files_in_main_branch":
		return w.ListFilesInMainBranch()

	case "list_files_in_bot_branch":
		return w.ListFilesInBotBranch()

	case "list_branches_in_repo":
		return w.ListBranchesInRepo()

	case "set_active_branch":
		return w.SetActiveBranch(query)

	case "create_branch":
		return w.CreateBranch(query)

	case "get_files_from_directory":
		return w.GetFilesFromDirectory(query)

	case "search_issues_and_prs":
		return w.SearchIssuesAndPRs(query)

	case "search_code":
		return w.SearchCode(query)

	case "get_latest_release":
		return w.GetLatestRelease()

	case "get_releases":
		return w.GetReleases()

	case "get_release":
		return w.GetRelease(query)

	default:
		return "", fmt.Errorf("invalid mode: %s", mode)
	}
}

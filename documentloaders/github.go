package documentloaders

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// GitHubIssuesLoader loads issues from a GitHub repository as documents.
type GitHubIssuesLoader struct {
	Repo         string   // Repository in format "owner/repo"
	AccessToken  string   // GitHub personal access token
	GitHubAPIURL string   // GitHub API URL, defaults to https://api.github.com
	IncludePRs   bool     // Include pull requests in results
	Milestone    *string  // Filter by milestone (number, "*" for any, "none" for no milestone)
	State        string   // Filter by state: "open", "closed", "all"
	Assignee     string   // Filter by assignee
	Creator      string   // Filter by creator
	Mentioned    string   // Filter by mentioned user
	Labels       []string // Filter by labels
	Sort         string   // Sort by: "created", "updated", "comments"
	Direction    string   // Sort direction: "asc", "desc"
	Since        string   // Only issues updated after this date (ISO 8601)
	Page         *int     // Page number for pagination
	PerPage      *int     // Items per page
}

var _ Loader = (*GitHubIssuesLoader)(nil)

// NewGitHubIssuesLoader creates a new GitHub issues loader.
func NewGitHubIssuesLoader(repo string, opts ...GitHubIssuesLoaderOption) (*GitHubIssuesLoader, error) {
	if repo == "" {
		return nil, errors.New("repository cannot be empty")
	}

	loader := &GitHubIssuesLoader{
		Repo:         repo,
		AccessToken:  os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"),
		GitHubAPIURL: "https://api.github.com",
		IncludePRs:   true,
		State:        "open",
	}

	if loader.AccessToken == "" {
		return nil, errors.New("GITHUB_PERSONAL_ACCESS_TOKEN environment variable is required")
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// GitHubIssuesLoaderOption is a function type for configuring GitHubIssuesLoader.
type GitHubIssuesLoaderOption func(*GitHubIssuesLoader)

// WithAccessToken sets the GitHub access token.
func WithAccessToken(token string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.AccessToken = token
	}
}

// WithIncludePRs sets whether to include pull requests.
func WithIncludePRs(include bool) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.IncludePRs = include
	}
}

// WithState sets the state filter.
func WithState(state string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.State = state
	}
}

// WithLabels sets the labels filter.
func WithLabels(labels []string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Labels = labels
	}
}

// WithMilestone sets the milestone filter.
func WithMilestone(milestone string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Milestone = &milestone
	}
}

// WithAssignee sets the assignee filter.
func WithAssignee(assignee string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Assignee = assignee
	}
}

// WithCreator sets the creator filter.
func WithCreator(creator string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Creator = creator
	}
}

// WithSort sets the sort field and direction.
func WithSort(sort, direction string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Sort = sort
		l.Direction = direction
	}
}

// WithSince sets the since filter.
func WithSince(since string) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Since = since
	}
}

// WithPagination sets pagination parameters.
func WithPagination(page, perPage int) GitHubIssuesLoaderOption {
	return func(l *GitHubIssuesLoader) {
		l.Page = &page
		l.PerPage = &perPage
	}
}

// Load loads GitHub issues as documents.
func (l *GitHubIssuesLoader) Load(ctx context.Context) ([]schema.Document, error) {
	var allDocs []schema.Document
	url := l.buildURL()

	client := &http.Client{Timeout: 30 * time.Second}

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+l.AccessToken)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		var issues []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, issue := range issues {
			doc := l.parseIssue(issue)
			if !l.IncludePRs && doc.Metadata["is_pull_request"].(bool) {
				continue
			}
			allDocs = append(allDocs, doc)
		}

		// Handle pagination
		if l.Page != nil || l.PerPage != nil {
			break // If specific pagination is set, don't auto-paginate
		}

		url = l.getNextURL(resp.Header.Get("Link"))
	}

	return allDocs, nil
}

// LoadAndSplit loads GitHub issues and splits them using a text splitter.
func (l *GitHubIssuesLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

func (l *GitHubIssuesLoader) buildURL() string {
	baseURL := fmt.Sprintf("%s/repos/%s/issues", l.GitHubAPIURL, l.Repo)
	params := url.Values{}

	if l.Milestone != nil {
		params.Add("milestone", *l.Milestone)
	}
	if l.State != "" {
		params.Add("state", l.State)
	}
	if l.Assignee != "" {
		params.Add("assignee", l.Assignee)
	}
	if l.Creator != "" {
		params.Add("creator", l.Creator)
	}
	if l.Mentioned != "" {
		params.Add("mentioned", l.Mentioned)
	}
	if len(l.Labels) > 0 {
		params.Add("labels", strings.Join(l.Labels, ","))
	}
	if l.Sort != "" {
		params.Add("sort", l.Sort)
	}
	if l.Direction != "" {
		params.Add("direction", l.Direction)
	}
	if l.Since != "" {
		params.Add("since", l.Since)
	}
	if l.Page != nil {
		params.Add("page", strconv.Itoa(*l.Page))
	}
	if l.PerPage != nil {
		params.Add("per_page", strconv.Itoa(*l.PerPage))
	}

	if len(params) > 0 {
		return baseURL + "?" + params.Encode()
	}
	return baseURL
}

func (l *GitHubIssuesLoader) parseIssue(issue map[string]interface{}) schema.Document {
	metadata := map[string]interface{}{
		"url":             getString(issue, "html_url"),
		"title":           getString(issue, "title"),
		"creator":         getNestedString(issue, "user", "login"),
		"created_at":      getString(issue, "created_at"),
		"comments":        getFloat64(issue, "comments"),
		"state":           getString(issue, "state"),
		"labels":          extractLabels(issue),
		"assignee":        getAssignee(issue),
		"milestone":       getMilestone(issue),
		"locked":          getBool(issue, "locked"),
		"number":          getFloat64(issue, "number"),
		"is_pull_request": issue["pull_request"] != nil,
	}

	content := getString(issue, "body")
	if content == "" {
		content = getString(issue, "title") // Use title if body is empty
	}

	return schema.Document{
		PageContent: content,
		Metadata:    metadata,
	}
}

func (l *GitHubIssuesLoader) getNextURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) == 2 && strings.Contains(parts[1], `rel="next"`) {
			url := strings.Trim(strings.TrimSpace(parts[0]), "<>")
			return url
		}
	}
	return ""
}

// GitHubFileLoader loads files from a GitHub repository as documents.
type GitHubFileLoader struct {
	Repo         string            // Repository in format "owner/repo"
	AccessToken  string            // GitHub personal access token
	GitHubAPIURL string            // GitHub API URL, defaults to https://api.github.com
	Branch       string            // Branch to load files from
	FileFilter   func(string) bool // Optional filter function for file paths
}

var _ Loader = (*GitHubFileLoader)(nil)

// NewGitHubFileLoader creates a new GitHub file loader.
func NewGitHubFileLoader(repo string, opts ...GitHubFileLoaderOption) (*GitHubFileLoader, error) {
	if repo == "" {
		return nil, errors.New("repository cannot be empty")
	}

	loader := &GitHubFileLoader{
		Repo:         repo,
		AccessToken:  os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"),
		GitHubAPIURL: "https://api.github.com",
		Branch:       "main",
	}

	if loader.AccessToken == "" {
		return nil, errors.New("GITHUB_PERSONAL_ACCESS_TOKEN environment variable is required")
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// GitHubFileLoaderOption is a function type for configuring GitHubFileLoader.
type GitHubFileLoaderOption func(*GitHubFileLoader)

// WithFileAccessToken sets the GitHub access token.
func WithFileAccessToken(token string) GitHubFileLoaderOption {
	return func(l *GitHubFileLoader) {
		l.AccessToken = token
	}
}

// WithBranch sets the branch to load files from.
func WithBranch(branch string) GitHubFileLoaderOption {
	return func(l *GitHubFileLoader) {
		l.Branch = branch
	}
}

// WithFileFilter sets a filter function for file paths.
func WithFileFilter(filter func(string) bool) GitHubFileLoaderOption {
	return func(l *GitHubFileLoader) {
		l.FileFilter = filter
	}
}

// Load loads GitHub files as documents.
func (l *GitHubFileLoader) Load(ctx context.Context) ([]schema.Document, error) {
	files, err := l.getFilePaths(ctx)
	if err != nil {
		return nil, err
	}

	var docs []schema.Document
	client := &http.Client{Timeout: 30 * time.Second}

	for _, file := range files {
		if file["type"] != "blob" { // Only process files, not directories
			continue
		}

		path := file["path"].(string)
		if l.FileFilter != nil && !l.FileFilter(path) {
			continue
		}

		content, err := l.getFileContent(ctx, client, path)
		if err != nil {
			continue // Skip files that can't be loaded
		}

		if content == "" {
			continue // Skip empty files
		}

		metadata := map[string]interface{}{
			"path":   path,
			"sha":    file["sha"],
			"source": fmt.Sprintf("%s/%s/%s/%s/%s", l.GitHubAPIURL, l.Repo, file["type"], l.Branch, path),
		}

		docs = append(docs, schema.Document{
			PageContent: content,
			Metadata:    metadata,
		})
	}

	return docs, nil
}

// LoadAndSplit loads GitHub files and splits them using a text splitter.
func (l *GitHubFileLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

func (l *GitHubFileLoader) getFilePaths(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/git/trees/%s?recursive=1", l.GitHubAPIURL, l.Repo, l.Branch)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+l.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file tree: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tree, ok := result["tree"].([]interface{})
	if !ok {
		return nil, errors.New("invalid tree response format")
	}

	var files []map[string]interface{}
	for _, item := range tree {
		if file, ok := item.(map[string]interface{}); ok {
			files = append(files, file)
		}
	}

	return files, nil
}

func (l *GitHubFileLoader) getFileContent(ctx context.Context, client *http.Client, path string) (string, error) {
	queryParams := ""
	if l.Branch != "" {
		queryParams = "?ref=" + l.Branch
	}

	url := fmt.Sprintf("%s/repos/%s/contents/%s%s", l.GitHubAPIURL, l.Repo, path, queryParams)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+l.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error for file %s: %s", path, resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode file content response: %w", err)
	}

	contentEncoded, ok := result["content"].(string)
	if !ok {
		return "", errors.New("no content field in response")
	}

	// Remove newlines from base64 encoded content
	contentEncoded = strings.ReplaceAll(contentEncoded, "\n", "")
	contentBytes, err := base64.StdEncoding.DecodeString(contentEncoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return string(contentBytes), nil
}

// Helper functions for parsing issue data
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getNestedString(data map[string]interface{}, parentKey, childKey string) string {
	if parent, ok := data[parentKey].(map[string]interface{}); ok {
		return getString(parent, childKey)
	}
	return ""
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func extractLabels(issue map[string]interface{}) []string {
	if labels, ok := issue["labels"].([]interface{}); ok {
		var labelNames []string
		for _, label := range labels {
			if labelMap, ok := label.(map[string]interface{}); ok {
				if name := getString(labelMap, "name"); name != "" {
					labelNames = append(labelNames, name)
				}
			}
		}
		return labelNames
	}
	return []string{}
}

func getAssignee(issue map[string]interface{}) string {
	if assignee, ok := issue["assignee"].(map[string]interface{}); ok && assignee != nil {
		return getString(assignee, "login")
	}
	return ""
}

func getMilestone(issue map[string]interface{}) string {
	if milestone, ok := issue["milestone"].(map[string]interface{}); ok && milestone != nil {
		return getString(milestone, "title")
	}
	return ""
}

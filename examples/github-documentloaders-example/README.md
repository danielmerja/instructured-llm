# GitHub Document Loaders Example

This example demonstrates how to use the GitHub document loaders from LangChain Go to load issues and files from GitHub repositories as documents.

## Overview

The GitHub document loaders provide a way to load content from GitHub repositories into LangChain documents, which can then be used for various NLP tasks, embeddings, or RAG (Retrieval-Augmented Generation) applications.

### Available Loaders

1. **`GitHubIssuesLoader`** - Loads GitHub issues and pull requests as documents
2. **`GitHubFileLoader`** - Loads files from a GitHub repository as documents

## Prerequisites

Before running this example, you need:

1. **GitHub Personal Access Token** - Required for authenticating with the GitHub API
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Generate a new token (classic) with appropriate scopes
   - For public repositories: `public_repo` scope
   - For private repositories: `repo` scope

2. Set the environment variable:
   ```bash
   export GITHUB_PERSONAL_ACCESS_TOKEN="your_personal_access_token_here"
   ```

## Running the Example

```bash
go run main.go
```

## What the Example Does

The example demonstrates three main use cases:

### 1. Loading GitHub Issues
```go
loader, err := documentloaders.NewGitHubIssuesLoader("owner/repo",
    documentloaders.WithAccessToken(token),
    documentloaders.WithState("all"), // Load both open and closed issues
    documentloaders.WithIncludePRs(true), // Include pull requests
    documentloaders.WithPagination(1, 5), // Limit results
)

docs, err := loader.Load(context.Background())
```

**Features:**
- Load both issues and pull requests
- Filter by state (open, closed, all)
- Filter by labels, assignee, creator, milestone
- Pagination support
- Rich metadata including title, creator, state, labels, etc.

### 2. Loading GitHub Files
```go
loader, err := documentloaders.NewGitHubFileLoader("owner/repo",
    documentloaders.WithFileAccessToken(token),
    documentloaders.WithBranch("main"),
    documentloaders.WithFileFilter(func(path string) bool {
        return strings.HasSuffix(path, ".md")
    }),
)

docs, err := loader.Load(context.Background())
```

**Features:**
- Load files from specific branches
- Filter files by path using custom functions
- Supports all text-based file types
- Metadata includes file path, SHA, and source URL

### 3. Load and Split Documents
```go
splitter := textsplitter.NewRecursiveCharacter()
splitter.ChunkSize = 500
splitter.ChunkOverlap = 50

splitDocs, err := loader.LoadAndSplit(context.Background(), splitter)
```

**Features:**
- Automatically split large documents into smaller chunks
- Configurable chunk size and overlap
- Useful for embedding large documents or RAG applications

## Configuration Options

### GitHubIssuesLoader Options

```go
loader, err := documentloaders.NewGitHubIssuesLoader("owner/repo",
    // Authentication
    documentloaders.WithAccessToken("your-token"),
    
    // Filtering
    documentloaders.WithState("open"),        // "open", "closed", "all"
    documentloaders.WithIncludePRs(false),    // Include pull requests
    documentloaders.WithLabels([]string{"bug", "enhancement"}),
    documentloaders.WithMilestone("v1.0"),    // Milestone title or number
    documentloaders.WithAssignee("username"), // Filter by assignee
    documentloaders.WithCreator("username"),  // Filter by creator
    
    // Sorting
    documentloaders.WithSort("updated", "desc"), // Sort by: created, updated, comments
    
    // Date filtering
    documentloaders.WithSince("2023-01-01T00:00:00Z"), // ISO 8601 format
    
    // Pagination
    documentloaders.WithPagination(1, 50), // page, per_page
)
```

### GitHubFileLoader Options

```go
loader, err := documentloaders.NewGitHubFileLoader("owner/repo",
    // Authentication
    documentloaders.WithFileAccessToken("your-token"),
    
    // Branch selection
    documentloaders.WithBranch("develop"),
    
    // File filtering
    documentloaders.WithFileFilter(func(path string) bool {
        // Only load Go files
        return strings.HasSuffix(path, ".go")
    }),
)
```

## Document Structure

### Issue Documents
Each issue/PR is loaded as a document with:

**PageContent**: The issue body text
**Metadata**:
- `url`: GitHub URL of the issue
- `title`: Issue title
- `creator`: Username of the creator
- `created_at`: Creation timestamp
- `state`: "open" or "closed"
- `labels`: Array of label names
- `assignee`: Assigned user (if any)
- `milestone`: Milestone title (if any)
- `number`: Issue/PR number
- `is_pull_request`: Boolean indicating if it's a PR
- `comments`: Number of comments
- `locked`: Boolean indicating if locked

### File Documents
Each file is loaded as a document with:

**PageContent**: The file content (decoded from base64)
**Metadata**:
- `path`: File path in the repository
- `sha`: Git SHA of the file
- `source`: Source URL

## Use Cases

1. **Documentation Search**: Load all markdown files from a repository for documentation search
2. **Issue Analysis**: Analyze issues and PRs for project insights
3. **Code Documentation**: Load code files for automated documentation generation
4. **RAG Applications**: Use repository content to build context-aware chatbots
5. **Project Analytics**: Analyze issue patterns, labels, and activity

## Advanced Usage

### Custom File Filters
```go
// Only load specific file types
documentloaders.WithFileFilter(func(path string) bool {
    return strings.HasSuffix(path, ".go") || 
           strings.HasSuffix(path, ".md") ||
           strings.HasSuffix(path, ".py")
})

// Exclude certain directories
documentloaders.WithFileFilter(func(path string) bool {
    return !strings.Contains(path, "vendor/") &&
           !strings.Contains(path, "node_modules/")
})
```

### Loading Multiple Repositories
```go
repos := []string{"owner/repo1", "owner/repo2", "owner/repo3"}
var allDocs []schema.Document

for _, repo := range repos {
    loader, _ := documentloaders.NewGitHubIssuesLoader(repo, 
        documentloaders.WithAccessToken(token))
    docs, _ := loader.Load(context.Background())
    allDocs = append(allDocs, docs...)
}
```

## Error Handling

The loaders include comprehensive error handling for:
- Authentication failures
- Repository not found
- Rate limiting
- Network timeouts
- Invalid file formats

## Rate Limiting

GitHub API has rate limits:
- **Authenticated requests**: 5,000 requests per hour
- **Unauthenticated requests**: 60 requests per hour

The loaders respect these limits but don't implement automatic retry logic. For production use, consider implementing rate limiting and retry mechanisms.

## Security

- Never commit your GitHub token to version control
- Use environment variables or secure configuration management
- Consider using GitHub Apps for production applications with higher rate limits 
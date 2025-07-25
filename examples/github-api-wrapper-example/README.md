# GitHub API Wrapper Example

This example demonstrates how to use the GitHub API wrapper utility from LangChain Go. This wrapper provides a comprehensive interface to GitHub operations and is equivalent to the Python `GitHubAPIWrapper` from LangChain.

## Overview

The GitHub API wrapper provides:
- A unified interface for all GitHub operations
- Support for both direct method calls and run-based dispatch
- Branch management capabilities
- File operations (create, read, update, delete)
- Issue and pull request management
- Search functionality
- Release management

## Prerequisites

This wrapper uses GitHub App authentication, so you need:

1. **GitHub App** - Create a GitHub App in your repository settings
   - Go to Settings > Developer settings > GitHub Apps
   - Create a new GitHub App
   - Note down the App ID
   - Generate and download a private key

2. **Environment Variables**:
   ```bash
   export GITHUB_REPOSITORY="owner/repo-name"
   export GITHUB_APP_ID="123456"
   export GITHUB_APP_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n..."
   ```

   Or if using a private key file:
   ```bash
   export GITHUB_REPOSITORY="owner/repo-name"
   export GITHUB_APP_ID="123456"
   export GITHUB_APP_PRIVATE_KEY="/path/to/private-key.pem"
   ```

3. **App Installation** - Install the GitHub App on your repository
   - Go to the app settings and install it on the target repository

## Running the Example

```bash
go run main.go
```

## Features Demonstrated

### 1. Run Method (Python-Compatible Interface)
The wrapper provides a `Run(mode, query)` method that matches the Python version:

```go
// Get issues
result, err := wrapper.Run("get_issues", "")

// Get specific issue
result, err := wrapper.Run("get_issue", "42")

// Create a file
result, err := wrapper.Run("create_file", "test.md\n\n# Test File")

// Update a file
fileUpdate := `test.md

OLD <<<<
# Test File
>>>> OLD
NEW <<<<
# Updated Test File
>>>> NEW`
result, err := wrapper.Run("update_file", fileUpdate)
```

### 2. Direct Method Calls
For better type safety and IDE support, you can use direct methods:

```go
// Issue operations
issues, err := wrapper.GetIssues()
issue, err := wrapper.GetIssue(42)
result, err := wrapper.CommentOnIssue("42\n\nGreat work!")

// Pull request operations
prs, err := wrapper.ListOpenPullRequests()
pr, err := wrapper.GetPullRequest(10)
result, err := wrapper.CreatePullRequest("New Feature\n\nAdded amazing functionality")

// File operations
content, err := wrapper.ReadFile("README.md")
result, err := wrapper.CreateFile("newfile.txt\n\nFile content")
result, err := wrapper.UpdateFile(updateString)
result, err := wrapper.DeleteFile("oldfile.txt")

// Branch operations
branches, err := wrapper.ListBranchesInRepo()
result, err := wrapper.SetActiveBranch("develop")
result, err := wrapper.CreateBranch("feature/new-feature")

// Search operations
results, err := wrapper.SearchIssuesAndPRs("bug")
results, err := wrapper.SearchCode("function main")

// Release operations
latest, err := wrapper.GetLatestRelease()
releases, err := wrapper.GetReleases()
release, err := wrapper.GetRelease("v1.0.0")
```

### 3. Configuration Options

```go
// Basic configuration (reads from environment)
wrapper, err := githubutil.NewGitHubAPIWrapper(nil)

// Custom configuration
wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
    Repository:       "owner/repo",
    AppID:           "123456",
    PrivateKey:      "-----BEGIN RSA PRIVATE KEY-----\n...",
    ActiveBranch:    "develop",
    GitHubBaseBranch: "main",
})
```

## Available Run Modes

The wrapper supports all the same modes as the Python version:

**Issue Operations:**
- `get_issues` - List open issues
- `get_issue` - Get specific issue by number
- `comment_on_issue` - Add comment to issue

**Pull Request Operations:**
- `list_open_pull_requests` - List open PRs
- `get_pull_request` - Get specific PR by number
- `create_pull_request` - Create new PR

**File Operations:**
- `read_file` - Read file content
- `create_file` - Create new file
- `update_file` - Update existing file
- `delete_file` - Delete file

**Branch Operations:**
- `list_branches_in_repo` - List all branches
- `set_active_branch` - Switch active branch
- `create_branch` - Create new branch

**Directory Operations:**
- `list_files_in_main_branch` - List files in main branch
- `list_files_in_bot_branch` - List files in active branch
- `get_files_from_directory` - List files in specific directory

**Search Operations:**
- `search_issues_and_prs` - Search issues and PRs
- `search_code` - Search code in repository

**Release Operations:**
- `get_latest_release` - Get latest release
- `get_releases` - Get recent releases
- `get_release` - Get specific release by tag

## Error Handling

The wrapper provides comprehensive error handling:

```go
result, err := wrapper.Run("get_issue", "invalid")
if err != nil {
    // Handle parsing errors, API errors, etc.
    log.Printf("Operation failed: %v", err)
}

// Many operations return user-friendly error messages in the result string
// rather than Go errors, matching the Python behavior
if strings.Contains(result, "Error") {
    log.Printf("GitHub operation message: %s", result)
}
```

## Authentication Notes

**Important:** This wrapper uses a simplified authentication mechanism. For production use, you should implement proper GitHub App authentication with:

1. JWT token generation for GitHub App authentication
2. Installation token retrieval and refresh
3. Proper token rotation and expiration handling

## Use Cases

1. **Automated Repository Management** - Create issues, PRs, manage branches
2. **Code Review Automation** - Comment on PRs, request reviews
3. **Documentation Updates** - Automatically update files and documentation
4. **Release Management** - Manage releases and changelogs
5. **Repository Analysis** - Search and analyze repository content
6. **CI/CD Integration** - Integrate with build and deployment pipelines

## Differences from Python Version

1. **Type Safety** - Go version provides better type safety
2. **Error Handling** - More explicit error handling patterns
3. **Configuration** - Struct-based configuration instead of Pydantic models
4. **Authentication** - Simplified for this example (production needs proper JWT flow)
5. **Performance** - Generally faster execution and lower memory usage

## Security Considerations

- Never commit private keys to version control
- Use environment variables or secure configuration management
- Implement proper JWT authentication for production use
- Consider using GitHub's OIDC tokens in CI/CD environments
- Rotate private keys regularly 
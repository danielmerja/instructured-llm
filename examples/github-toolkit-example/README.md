# GitHub Toolkit Example

This example demonstrates how to use the GitHub toolkit from LangChain Go to interact with GitHub repositories.

## Overview

The GitHub toolkit provides a comprehensive set of tools for working with GitHub repositories, including:

- **Issue Management**: List issues, get specific issues, comment on issues
- **Pull Request Management**: List PRs, get PR details, create PRs, list PR files
- **File Operations**: Read, create, update, and delete files
- **Repository Navigation**: List branches, browse directories
- **Search**: Search code and issues/PRs
- **Release Management**: Get releases and release information (optional)

## Prerequisites

Before running this example, you need to set up the following environment variables:

1. **`GITHUB_TOKEN`**: Your GitHub personal access token
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Generate a new token with appropriate permissions for your repository
   - For public repositories, you need: `public_repo`, `read:org`, `read:user`
   - For private repositories, you need: `repo` scope

2. **`GITHUB_REPOSITORY`**: The repository you want to interact with
   - Format: `owner/repository-name`
   - Example: `tmc/langchaingo`

## Setting up Environment Variables

### On Unix/Linux/macOS:
```bash
export GITHUB_TOKEN="your_github_token_here"
export GITHUB_REPOSITORY="owner/repo-name"
```

### On Windows (Command Prompt):
```cmd
set GITHUB_TOKEN=your_github_token_here
set GITHUB_REPOSITORY=owner/repo-name
```

### On Windows (PowerShell):
```powershell
$env:GITHUB_TOKEN="your_github_token_here"
$env:GITHUB_REPOSITORY="owner/repo-name"
```

## Running the Example

Once you have set the environment variables, you can run the example:

```bash
go run main.go
```

## What the Example Does

The example demonstrates several common operations:

1. **Lists available tools** in the toolkit
2. **Lists repository issues** - shows open issues in the repository
3. **Lists repository branches** - displays all branches
4. **Reads a file** - attempts to read the README.md file
5. **Lists directory contents** - shows files in the root directory
6. **Lists pull requests** - displays open PRs
7. **Searches code** - searches for "func main" in the repository
8. **Gets latest release** - fetches the most recent release (if available)

## Using Individual Tools

You can also create and use individual tools instead of the full toolkit:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/tmc/langchaingo/tools/github"
)

func main() {
    // Create a single tool
    getIssuesTool, err := github.NewGetIssuesTool()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the tool
    result, err := getIssuesTool.Call(context.Background(), "")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result)
}
```

## Toolkit Options

The toolkit can be configured with various options:

```go
// Create toolkit with all tools including release tools
toolkit, err := github.NewToolkit(github.ToolkitOptions{
    IncludeReleaseTools: true,
})

// Create toolkit with default options (no release tools)
toolkit, err := github.NewToolkit()
```

## Available Tools

When `IncludeReleaseTools` is `false` (default):
- Get Issues
- Get Issue
- Comment on Issue
- List Pull Requests
- Get Pull Request
- Create Pull Request
- List Pull Request Files
- Read File
- Create File
- Update File
- Delete File
- List Branches
- Get Directory Files
- Search Code
- Search Issues and PRs

When `IncludeReleaseTools` is `true`, additional tools are included:
- Get Releases
- Get Latest Release
- Get Release

## Error Handling

The example includes basic error handling. In production code, you should handle errors appropriately for your use case. Common errors include:

- **Authentication errors**: Invalid or expired GitHub token
- **Permission errors**: Token doesn't have required permissions
- **Not found errors**: Repository, file, issue, or PR doesn't exist
- **Rate limiting**: GitHub API rate limits exceeded

## Security Note

**Never commit your GitHub token to version control.** Always use environment variables or secure configuration management for sensitive credentials. 
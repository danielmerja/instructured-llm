package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	githubutil "github.com/tmc/langchaingo/util/github"
)

func main() {
	// Check for required environment variables
	repo := os.Getenv("GITHUB_REPOSITORY")
	appID := os.Getenv("GITHUB_APP_ID")
	privateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")

	if repo == "" || appID == "" || privateKey == "" {
		log.Fatal("Required environment variables: GITHUB_REPOSITORY, GITHUB_APP_ID, GITHUB_APP_PRIVATE_KEY")
	}

	fmt.Printf("GitHub API Wrapper Example\n")
	fmt.Printf("Repository: %s\n\n", repo)

	// Create the GitHub API wrapper
	wrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: repo,
		AppID:      appID,
		PrivateKey: privateKey,
	})
	if err != nil {
		log.Fatalf("Failed to create GitHub API wrapper: %v", err)
	}

	// Example 1: Using the Run method (similar to Python version)
	fmt.Println("=== Example 1: Using Run method ===")
	demonstrateRunMethod(wrapper)
	fmt.Println()

	// Example 2: Using direct methods
	fmt.Println("=== Example 2: Using direct methods ===")
	demonstrateDirectMethods(wrapper)
	fmt.Println()

	// Example 3: File operations
	fmt.Println("=== Example 3: File operations ===")
	demonstrateFileOperations(wrapper)
	fmt.Println()

	// Example 4: Branch operations
	fmt.Println("=== Example 4: Branch operations ===")
	demonstrateBranchOperations(wrapper)

	fmt.Println("\nGitHub API Wrapper Example completed!")
}

func demonstrateRunMethod(wrapper *githubutil.GitHubAPIWrapper) {
	// Get issues using the run method
	result, err := wrapper.Run("get_issues", "")
	if err != nil {
		fmt.Printf("Error getting issues: %v\n", err)
	} else {
		fmt.Printf("Issues: %s\n", result)
	}

	// List branches using the run method
	result, err = wrapper.Run("list_branches_in_repo", "")
	if err != nil {
		fmt.Printf("Error listing branches: %v\n", err)
	} else {
		fmt.Printf("Branches: %s\n", result)
	}

	// List files in main branch
	result, err = wrapper.Run("list_files_in_main_branch", "")
	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
	} else {
		fmt.Printf("Files in main branch: %s\n", result)
	}
}

func demonstrateDirectMethods(wrapper *githubutil.GitHubAPIWrapper) {
	// Get open pull requests
	result, err := wrapper.ListOpenPullRequests()
	if err != nil {
		fmt.Printf("Error getting PRs: %v\n", err)
	} else {
		fmt.Printf("Open PRs: %s\n", result)
	}

	// Get latest release
	result, err = wrapper.GetLatestRelease()
	if err != nil {
		fmt.Printf("Error getting latest release: %v\n", err)
	} else {
		fmt.Printf("Latest release: %s\n", result)
	}

	// Search issues and PRs
	result, err = wrapper.SearchIssuesAndPRs("bug")
	if err != nil {
		fmt.Printf("Error searching issues: %v\n", err)
	} else {
		fmt.Printf("Search results: %s\n", result)
	}
}

func demonstrateFileOperations(wrapper *githubutil.GitHubAPIWrapper) {
	// Read README file
	content, err := wrapper.ReadFile("README.md")
	if err != nil {
		fmt.Printf("Error reading README: %v\n", err)
	} else {
		// Truncate for display
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("README.md content: %s\n", content)
	}

	// List files from root directory
	result, err := wrapper.GetFilesFromDirectory("")
	if err != nil {
		fmt.Printf("Error listing directory files: %v\n", err)
	} else {
		// Show just first few files
		lines := strings.Split(result, "\n")
		if len(lines) > 5 {
			lines = lines[:5]
			result = strings.Join(lines, "\n") + "\n... (truncated)"
		}
		fmt.Printf("Root directory files: %s\n", result)
	}
}

func demonstrateBranchOperations(wrapper *githubutil.GitHubAPIWrapper) {
	// List all branches
	result, err := wrapper.ListBranchesInRepo()
	if err != nil {
		fmt.Printf("Error listing branches: %v\n", err)
	} else {
		fmt.Printf("All branches: %s\n", result)
	}

	// Try to set active branch to main
	result, err = wrapper.SetActiveBranch("main")
	if err != nil {
		fmt.Printf("Error setting active branch: %v\n", err)
	} else {
		fmt.Printf("Set active branch result: %s\n", result)
	}

	// List files in current active branch
	result, err = wrapper.ListFilesInBotBranch()
	if err != nil {
		fmt.Printf("Error listing bot branch files: %v\n", err)
	} else {
		// Show just first few files
		lines := strings.Split(result, "\n")
		if len(lines) > 8 {
			lines = lines[:8]
			result = strings.Join(lines, "\n") + "\n... (truncated)"
		}
		fmt.Printf("Files in active branch: %s\n", result)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/tools/github"
)

func main() {
	// Check for required environment variables
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY")

	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	if repo == "" {
		log.Fatal("GITHUB_REPOSITORY environment variable is required (format: owner/repo)")
	}

	fmt.Printf("GitHub Toolkit Example\n")
	fmt.Printf("Repository: %s\n\n", repo)

	// Create the GitHub toolkit with all tools
	toolkit, err := github.NewToolkit(github.ToolkitOptions{
		IncludeReleaseTools: true, // Include release tools as well
	})
	if err != nil {
		log.Fatalf("Failed to create GitHub toolkit: %v", err)
	}

	// Display available tools
	fmt.Println("Available GitHub Tools:")
	for i, name := range toolkit.GetToolNames() {
		fmt.Printf("%d. %s\n", i+1, name)
	}
	fmt.Println()

	// Example 1: List repository issues
	fmt.Println("=== Example 1: List Repository Issues ===")
	getIssuesTool := toolkit.GetToolByName("Get Issues")
	if getIssuesTool != nil {
		result, err := getIssuesTool.Call(context.Background(), "")
		if err != nil {
			fmt.Printf("Error listing issues: %v\n", err)
		} else {
			fmt.Println(result)
		}
	}
	fmt.Println()

	// Example 2: List repository branches
	fmt.Println("=== Example 2: List Repository Branches ===")
	listBranchesTool := toolkit.GetToolByName("List Branches")
	if listBranchesTool != nil {
		result, err := listBranchesTool.Call(context.Background(), "")
		if err != nil {
			fmt.Printf("Error listing branches: %v\n", err)
		} else {
			fmt.Println(result)
		}
	}
	fmt.Println()

	// Example 3: Read a specific file (README.md)
	fmt.Println("=== Example 3: Read README.md File ===")
	readFileTool := toolkit.GetToolByName("Read File")
	if readFileTool != nil {
		result, err := readFileTool.Call(context.Background(), "README.md")
		if err != nil {
			fmt.Printf("Error reading README.md: %v\n", err)
		} else {
			// Truncate output for readability in example
			if len(result) > 500 {
				fmt.Printf("%s...\n[truncated for example]\n", result[:500])
			} else {
				fmt.Println(result)
			}
		}
	}
	fmt.Println()

	// Example 4: List files in root directory
	fmt.Println("=== Example 4: List Files in Root Directory ===")
	getDirectoryFilesTool := toolkit.GetToolByName("Get Directory Files")
	if getDirectoryFilesTool != nil {
		result, err := getDirectoryFilesTool.Call(context.Background(), "")
		if err != nil {
			fmt.Printf("Error listing directory files: %v\n", err)
		} else {
			fmt.Println(result)
		}
	}
	fmt.Println()

	// Example 5: List pull requests
	fmt.Println("=== Example 5: List Pull Requests ===")
	listPRsTool := toolkit.GetToolByName("List Pull Requests")
	if listPRsTool != nil {
		result, err := listPRsTool.Call(context.Background(), "")
		if err != nil {
			fmt.Printf("Error listing pull requests: %v\n", err)
		} else {
			fmt.Println(result)
		}
	}
	fmt.Println()

	// Example 6: Search for code
	fmt.Println("=== Example 6: Search for Code ===")
	searchCodeTool := toolkit.GetToolByName("Search Code")
	if searchCodeTool != nil {
		result, err := searchCodeTool.Call(context.Background(), "func main")
		if err != nil {
			fmt.Printf("Error searching code: %v\n", err)
		} else {
			fmt.Println(result)
		}
	}
	fmt.Println()

	// Example 7: Get latest release (if release tools are enabled)
	fmt.Println("=== Example 7: Get Latest Release ===")
	getLatestReleaseTool := toolkit.GetToolByName("Get Latest Release")
	if getLatestReleaseTool != nil {
		result, err := getLatestReleaseTool.Call(context.Background(), "")
		if err != nil {
			fmt.Printf("Error getting latest release: %v\n", err)
		} else {
			fmt.Println(result)
		}
	} else {
		fmt.Println("Latest release tool not available (release tools not enabled)")
	}

	fmt.Println("\nGitHub Toolkit Example completed!")
}

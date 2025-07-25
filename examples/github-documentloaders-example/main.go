package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

func main() {
	// Check for required environment variable
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_PERSONAL_ACCESS_TOKEN environment variable is required")
	}

	// Example repository - you can change this to any public repository
	repo := "octocat/Hello-World"
	fmt.Printf("GitHub Document Loaders Example\n")
	fmt.Printf("Repository: %s\n\n", repo)

	// Example 1: Load GitHub Issues
	fmt.Println("=== Example 1: Loading GitHub Issues ===")
	loadGitHubIssues(repo, token)
	fmt.Println()

	// Example 2: Load GitHub Files
	fmt.Println("=== Example 2: Loading GitHub Files ===")
	loadGitHubFiles(repo, token)
	fmt.Println()

	// Example 3: Load and Split Documents
	fmt.Println("=== Example 3: Load and Split Documents ===")
	loadAndSplitIssues(repo, token)
	fmt.Println()

	fmt.Println("GitHub Document Loaders Example completed!")
}

func loadGitHubIssues(repo, token string) {
	// Create a GitHub issues loader with various options
	loader, err := documentloaders.NewGitHubIssuesLoader(repo,
		documentloaders.WithAccessToken(token),
		documentloaders.WithState("all"),     // Load both open and closed issues
		documentloaders.WithIncludePRs(true), // Include pull requests
		documentloaders.WithPagination(1, 5), // Limit to 5 issues for the example
	)
	if err != nil {
		log.Printf("Failed to create GitHub issues loader: %v", err)
		return
	}

	// Load the issues
	docs, err := loader.Load(context.Background())
	if err != nil {
		log.Printf("Failed to load issues: %v", err)
		return
	}

	if len(docs) == 0 {
		fmt.Println("No issues found in the repository.")
		return
	}

	fmt.Printf("Loaded %d issues/PRs:\n", len(docs))
	for i, doc := range docs {
		title := doc.Metadata["title"].(string)
		state := doc.Metadata["state"].(string)
		number := doc.Metadata["number"].(float64)
		isPR := doc.Metadata["is_pull_request"].(bool)

		itemType := "Issue"
		if isPR {
			itemType = "PR"
		}

		fmt.Printf("%d. %s #%.0f: %s [%s]\n", i+1, itemType, number, title, state)

		// Show a snippet of the content
		content := doc.PageContent
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		if content != "" {
			fmt.Printf("   Content: %s\n", content)
		}
		fmt.Println()
	}
}

func loadGitHubFiles(repo, token string) {
	// Create a GitHub file loader with a filter for specific file types
	loader, err := documentloaders.NewGitHubFileLoader(repo,
		documentloaders.WithFileAccessToken(token),
		documentloaders.WithBranch("main"), // Load from main branch
		documentloaders.WithFileFilter(func(path string) bool {
			// Only load markdown and text files for the example
			ext := strings.ToLower(path)
			return strings.HasSuffix(ext, ".md") ||
				strings.HasSuffix(ext, ".txt") ||
				strings.HasSuffix(ext, ".rst")
		}),
	)
	if err != nil {
		log.Printf("Failed to create GitHub file loader: %v", err)
		return
	}

	// Load the files
	docs, err := loader.Load(context.Background())
	if err != nil {
		log.Printf("Failed to load files: %v", err)
		return
	}

	if len(docs) == 0 {
		fmt.Println("No markdown/text files found in the repository.")
		return
	}

	fmt.Printf("Loaded %d files:\n", len(docs))
	for i, doc := range docs {
		path := doc.Metadata["path"].(string)
		fmt.Printf("%d. %s\n", i+1, path)

		// Show a snippet of the content
		content := doc.PageContent
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("   Content preview: %s\n", content)
		fmt.Println()
	}
}

func loadAndSplitIssues(repo, token string) {
	// Create a GitHub issues loader
	loader, err := documentloaders.NewGitHubIssuesLoader(repo,
		documentloaders.WithAccessToken(token),
		documentloaders.WithState("all"),
		documentloaders.WithPagination(1, 3), // Limit for example
	)
	if err != nil {
		log.Printf("Failed to create GitHub issues loader: %v", err)
		return
	}

	// Create a text splitter to split documents into smaller chunks
	splitter := textsplitter.NewRecursiveCharacter()
	splitter.ChunkSize = 500
	splitter.ChunkOverlap = 50

	// Load and split the documents
	splitDocs, err := loader.LoadAndSplit(context.Background(), splitter)
	if err != nil {
		log.Printf("Failed to load and split issues: %v", err)
		return
	}

	if len(splitDocs) == 0 {
		fmt.Println("No issues found to split.")
		return
	}

	fmt.Printf("Loaded and split into %d document chunks:\n", len(splitDocs))
	for i, doc := range splitDocs {
		// Get the original issue title from metadata
		title := "Unknown"
		if titleVal, exists := doc.Metadata["title"]; exists {
			if titleStr, ok := titleVal.(string); ok {
				title = titleStr
			}
		}

		fmt.Printf("%d. Chunk from: %s\n", i+1, title)

		// Show the chunk content
		content := doc.PageContent
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		fmt.Printf("   Chunk: %s\n", content)
		fmt.Println()
	}
}

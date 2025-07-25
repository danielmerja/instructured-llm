// Package github provides tools for interacting with GitHub repositories.
//
// This package allows agents to interact with GitHub repositories through the
// go-github library, providing operations for issues, pull requests, files,
// branches, releases, and more.
//
// To use these tools, you must set the following environment variables:
//   - GITHUB_TOKEN: Your GitHub personal access token
//   - GITHUB_REPOSITORY: The repository in format "owner/repo"
//
// Example usage:
//
//	import "github.com/tmc/langchaingo/tools/github"
//
//	// Create a toolkit with all GitHub tools
//	toolkit, err := github.NewToolkit()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get all tools
//	tools := toolkit.GetTools()
//
//	// Or create individual tools
//	getIssue, err := github.NewGetIssueTool()
//	if err != nil {
//		log.Fatal(err)
//	}
package github

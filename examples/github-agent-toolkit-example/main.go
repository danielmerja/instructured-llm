package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	githubutil "github.com/tmc/langchaingo/util/github"
)

func main() {
	fmt.Println("GitHub Agent Toolkit Example")
	fmt.Println("============================")

	// Check for required environment variables
	if err := checkEnvironmentVariables(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create OpenAI LLM
	llm, err := openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	// Create GitHub API wrapper
	githubWrapper, err := githubutil.NewGitHubAPIWrapper(&githubutil.Config{
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		AppID:      os.Getenv("GITHUB_APP_ID"),
		PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create GitHub API wrapper: %v", err)
	}

	// Create GitHub Agent Toolkit
	githubToolkit := agents.NewGitHubAgentToolkit(githubWrapper, agents.GitHubAgentToolkitOptions{
		IncludeReleaseTools: true, // Include release tools for full functionality
	})

	fmt.Printf("Created GitHub toolkit with %d tools:\n", len(githubToolkit.GetTools()))
	for i, toolName := range githubToolkit.GetToolNames() {
		fmt.Printf("  %d. %s\n", i+1, toolName)
	}
	fmt.Println()

	// Example 1: MRKL Agent with GitHub tools
	fmt.Println("=== Example 1: MRKL Agent with GitHub Tools ===")
	if err := demonstrateMRKLAgent(ctx, llm, githubToolkit); err != nil {
		log.Printf("MRKL Agent example failed: %v", err)
	}
	fmt.Println()

	// Example 2: Conversational Agent with GitHub tools
	fmt.Println("=== Example 2: Conversational Agent with GitHub Tools ===")
	if err := demonstrateConversationalAgent(ctx, llm, githubToolkit); err != nil {
		log.Printf("Conversational Agent example failed: %v", err)
	}
	fmt.Println()

	// Example 3: Custom Agent workflow
	fmt.Println("=== Example 3: Custom Repository Analysis Workflow ===")
	if err := demonstrateRepositoryAnalysis(ctx, llm, githubToolkit); err != nil {
		log.Printf("Repository analysis example failed: %v", err)
	}
	fmt.Println()

	// Example 4: Issue Management workflow
	fmt.Println("=== Example 4: Issue Management Workflow ===")
	if err := demonstrateIssueManagement(ctx, llm, githubToolkit); err != nil {
		log.Printf("Issue management example failed: %v", err)
	}

	fmt.Println("\nGitHub Agent Toolkit examples completed!")
}

func checkEnvironmentVariables() error {
	required := []string{
		"GITHUB_REPOSITORY",
		"GITHUB_APP_ID",
		"GITHUB_APP_PRIVATE_KEY",
		"OPENAI_API_KEY",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func demonstrateMRKLAgent(ctx context.Context, llm *openai.LLM, githubToolkit *agents.GitHubAgentToolkit) error {
	// Get a subset of tools for the MRKL agent (to avoid overwhelming the LLM)
	selectedTools := []tools.Tool{
		githubToolkit.GetToolByName("Get Issues"),
		githubToolkit.GetToolByName("Read File"),
		githubToolkit.GetToolByName("List branches in this repository"),
		githubToolkit.GetToolByName("Search code"),
	}

	// Remove nil tools (in case some names don't match)
	var validTools []tools.Tool
	for _, tool := range selectedTools {
		if tool != nil {
			validTools = append(validTools, tool)
		}
	}

	if len(validTools) == 0 {
		return fmt.Errorf("no valid tools found for MRKL agent")
	}

	// Create MRKL agent
	agent := agents.NewOneShotAgent(llm, validTools,
		agents.WithMaxIterations(3),
	)

	executor := agents.NewExecutor(agent)

	// Ask the agent to analyze the repository
	query := "What can you tell me about this GitHub repository? Please check the issues, read the README file, and list the available branches."

	result, err := chains.Run(ctx, executor, query)
	if err != nil {
		return fmt.Errorf("MRKL agent execution failed: %w", err)
	}

	fmt.Printf("MRKL Agent Analysis:\n%s\n", result)
	return nil
}

func demonstrateConversationalAgent(ctx context.Context, llm *openai.LLM, githubToolkit *agents.GitHubAgentToolkit) error {
	// Use a broader set of tools for the conversational agent
	conversationalTools := []tools.Tool{
		githubToolkit.GetToolByName("Get Issues"),
		githubToolkit.GetToolByName("Get Issue"),
		githubToolkit.GetToolByName("List open pull requests (PRs)"),
		githubToolkit.GetToolByName("Read File"),
		githubToolkit.GetToolByName("Search issues and pull requests"),
	}

	// Remove nil tools
	var validTools []tools.Tool
	for _, tool := range conversationalTools {
		if tool != nil {
			validTools = append(validTools, tool)
		}
	}

	if len(validTools) == 0 {
		return fmt.Errorf("no valid tools found for conversational agent")
	}

	// Create conversational agent
	agent := agents.NewConversationalAgent(llm, validTools,
		agents.WithMaxIterations(3),
	)

	executor := agents.NewExecutor(agent)

	// Have a conversation about the repository
	query := "Hi! I'm interested in understanding the current state of this repository. Can you help me find any open issues or pull requests that might need attention?"

	result, err := chains.Run(ctx, executor, query)
	if err != nil {
		return fmt.Errorf("conversational agent execution failed: %w", err)
	}

	fmt.Printf("Conversational Agent Response:\n%s\n", result)
	return nil
}

func demonstrateRepositoryAnalysis(ctx context.Context, llm *openai.LLM, githubToolkit *agents.GitHubAgentToolkit) error {
	// Create a focused set of tools for repository analysis
	analysisTools := []tools.Tool{
		githubToolkit.GetToolByName("Overview of existing files in Main branch"),
		githubToolkit.GetToolByName("List branches in this repository"),
		githubToolkit.GetToolByName("Get latest release"),
		githubToolkit.GetToolByName("Search code"),
	}

	// Remove nil tools
	var validTools []tools.Tool
	for _, tool := range analysisTools {
		if tool != nil {
			validTools = append(validTools, tool)
		}
	}

	if len(validTools) == 0 {
		return fmt.Errorf("no valid tools found for repository analysis")
	}

	// Create MRKL agent with custom prompt
	agent := agents.NewOneShotAgent(llm, validTools,
		agents.WithMaxIterations(4),
		agents.WithPromptPrefix(`You are a repository analyst AI. Your job is to analyze GitHub repositories and provide insights about their structure, activity, and health.`),
	)

	executor := agents.NewExecutor(agent)

	query := "Please provide a comprehensive analysis of this repository including its structure, recent activity, and development status. Look at the file structure, branches, latest release, and any notable code patterns."

	result, err := chains.Run(ctx, executor, query)
	if err != nil {
		return fmt.Errorf("repository analysis failed: %w", err)
	}

	fmt.Printf("Repository Analysis:\n%s\n", result)
	return nil
}

func demonstrateIssueManagement(ctx context.Context, llm *openai.LLM, githubToolkit *agents.GitHubAgentToolkit) error {
	// Focus on issue-related tools
	issueTools := []tools.Tool{
		githubToolkit.GetToolByName("Get Issues"),
		githubToolkit.GetToolByName("Get Issue"),
		githubToolkit.GetToolByName("Search issues and pull requests"),
		// Note: Comment creation would require actual interaction, so we'll exclude it for this demo
	}

	// Remove nil tools
	var validTools []tools.Tool
	for _, tool := range issueTools {
		if tool != nil {
			validTools = append(validTools, tool)
		}
	}

	if len(validTools) == 0 {
		return fmt.Errorf("no valid tools found for issue management")
	}

	// Create conversational agent focused on issue management
	agent := agents.NewConversationalAgent(llm, validTools,
		agents.WithMaxIterations(3),
		agents.WithPromptPrefix(`You are an issue management assistant. Help users understand and manage GitHub issues in their repository.`),
	)

	executor := agents.NewExecutor(agent)

	query := "Can you help me understand what issues are currently open in this repository? If there are any, please tell me about the most recent or important ones."

	result, err := chains.Run(ctx, executor, query)
	if err != nil {
		return fmt.Errorf("issue management workflow failed: %w", err)
	}

	fmt.Printf("Issue Management Summary:\n%s\n", result)
	return nil
}

// Helper function to demonstrate direct tool usage (without agents)
func demonstrateDirectToolUsage(ctx context.Context, githubToolkit *agents.GitHubAgentToolkit) error {
	fmt.Println("=== Direct Tool Usage Examples ===")

	// Example: Get Issues
	getIssuesTool := githubToolkit.GetToolByName("Get Issues")
	if getIssuesTool != nil {
		result, err := getIssuesTool.Call(ctx, "")
		if err != nil {
			fmt.Printf("Get Issues failed: %v\n", err)
		} else {
			fmt.Printf("Issues: %s\n", result)
		}
	}

	// Example: Read README
	readFileTool := githubToolkit.GetToolByName("Read File")
	if readFileTool != nil {
		result, err := readFileTool.Call(ctx, "README.md")
		if err != nil {
			fmt.Printf("Read README failed: %v\n", err)
		} else {
			// Truncate for display
			if len(result) > 300 {
				result = result[:300] + "..."
			}
			fmt.Printf("README content: %s\n", result)
		}
	}

	// Example: List branches
	listBranchesTool := githubToolkit.GetToolByName("List branches in this repository")
	if listBranchesTool != nil {
		result, err := listBranchesTool.Call(ctx, "")
		if err != nil {
			fmt.Printf("List branches failed: %v\n", err)
		} else {
			fmt.Printf("Branches: %s\n", result)
		}
	}

	return nil
}

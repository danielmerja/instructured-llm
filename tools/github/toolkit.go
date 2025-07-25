package github

import (
	"fmt"

	"github.com/tmc/langchaingo/tools"
)

// Toolkit represents a collection of GitHub tools.
type Toolkit struct {
	tools []tools.Tool
}

// ToolkitOptions represents options for creating a GitHub toolkit.
type ToolkitOptions struct {
	IncludeReleaseTools bool
}

// NewToolkit creates a new GitHub toolkit with all available tools.
func NewToolkit(opts ...ToolkitOptions) (*Toolkit, error) {
	var options ToolkitOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	var tools []tools.Tool

	// Issue tools
	getIssues, err := NewGetIssuesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get issues tool: %w", err)
	}
	tools = append(tools, getIssues)

	getIssue, err := NewGetIssueTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get issue tool: %w", err)
	}
	tools = append(tools, getIssue)

	commentOnIssue, err := NewCommentOnIssueTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create comment on issue tool: %w", err)
	}
	tools = append(tools, commentOnIssue)

	// Pull request tools
	listPRs, err := NewListPullRequestsTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create list pull requests tool: %w", err)
	}
	tools = append(tools, listPRs)

	getPR, err := NewGetPullRequestTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get pull request tool: %w", err)
	}
	tools = append(tools, getPR)

	createPR, err := NewCreatePullRequestTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create create pull request tool: %w", err)
	}
	tools = append(tools, createPR)

	listPRFiles, err := NewListPullRequestFilesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create list pull request files tool: %w", err)
	}
	tools = append(tools, listPRFiles)

	// File operation tools
	readFile, err := NewReadFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create read file tool: %w", err)
	}
	tools = append(tools, readFile)

	createFile, err := NewCreateFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create create file tool: %w", err)
	}
	tools = append(tools, createFile)

	updateFile, err := NewUpdateFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create update file tool: %w", err)
	}
	tools = append(tools, updateFile)

	deleteFile, err := NewDeleteFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create delete file tool: %w", err)
	}
	tools = append(tools, deleteFile)

	// Repository and branch tools
	listBranches, err := NewListBranchesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create list branches tool: %w", err)
	}
	tools = append(tools, listBranches)

	getDirectoryFiles, err := NewGetDirectoryFilesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get directory files tool: %w", err)
	}
	tools = append(tools, getDirectoryFiles)

	// Search tools
	searchCode, err := NewSearchCodeTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create search code tool: %w", err)
	}
	tools = append(tools, searchCode)

	searchIssuesAndPRs, err := NewSearchIssuesAndPRsTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create search issues and PRs tool: %w", err)
	}
	tools = append(tools, searchIssuesAndPRs)

	// Release tools (optional)
	if options.IncludeReleaseTools {
		getReleases, err := NewGetReleasesTool()
		if err != nil {
			return nil, fmt.Errorf("failed to create get releases tool: %w", err)
		}
		tools = append(tools, getReleases)

		getLatestRelease, err := NewGetLatestReleaseTool()
		if err != nil {
			return nil, fmt.Errorf("failed to create get latest release tool: %w", err)
		}
		tools = append(tools, getLatestRelease)

		getRelease, err := NewGetReleaseTool()
		if err != nil {
			return nil, fmt.Errorf("failed to create get release tool: %w", err)
		}
		tools = append(tools, getRelease)
	}

	return &Toolkit{
		tools: tools,
	}, nil
}

// GetTools returns all tools in the toolkit.
func (tk *Toolkit) GetTools() []tools.Tool {
	return tk.tools
}

// GetToolByName returns a tool by its name, or nil if not found.
func (tk *Toolkit) GetToolByName(name string) tools.Tool {
	for _, tool := range tk.tools {
		if tool.Name() == name {
			return tool
		}
	}
	return nil
}

// GetToolNames returns the names of all tools in the toolkit.
func (tk *Toolkit) GetToolNames() []string {
	names := make([]string, len(tk.tools))
	for i, tool := range tk.tools {
		names[i] = tool.Name()
	}
	return names
}

package agents

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/tools"
	githubutil "github.com/tmc/langchaingo/util/github"
)

// GitHubAgentToolkit provides GitHub operations for agents.
type GitHubAgentToolkit struct {
	wrapper             *githubutil.GitHubAPIWrapper
	includeReleaseTools bool
	tools               []tools.Tool
}

// GitHubAgentToolkitOptions holds configuration options for the toolkit.
type GitHubAgentToolkitOptions struct {
	IncludeReleaseTools bool
}

// NewGitHubAgentToolkit creates a new GitHub agent toolkit.
func NewGitHubAgentToolkit(wrapper *githubutil.GitHubAPIWrapper, opts ...GitHubAgentToolkitOptions) *GitHubAgentToolkit {
	var options GitHubAgentToolkitOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	toolkit := &GitHubAgentToolkit{
		wrapper:             wrapper,
		includeReleaseTools: options.IncludeReleaseTools,
	}

	toolkit.tools = toolkit.createTools()
	return toolkit
}

// FromGitHubAPIWrapper creates a GitHub agent toolkit from a GitHub API wrapper.
func FromGitHubAPIWrapper(wrapper *githubutil.GitHubAPIWrapper, includeReleaseTools bool) *GitHubAgentToolkit {
	return NewGitHubAgentToolkit(wrapper, GitHubAgentToolkitOptions{
		IncludeReleaseTools: includeReleaseTools,
	})
}

// GetTools returns all tools in the toolkit.
func (t *GitHubAgentToolkit) GetTools() []tools.Tool {
	return t.tools
}

// GetToolByName returns a tool by its name, or nil if not found.
func (t *GitHubAgentToolkit) GetToolByName(name string) tools.Tool {
	for _, tool := range t.tools {
		if tool.Name() == name {
			return tool
		}
	}
	return nil
}

// GetToolNames returns the names of all tools in the toolkit.
func (t *GitHubAgentToolkit) GetToolNames() []string {
	names := make([]string, len(t.tools))
	for i, tool := range t.tools {
		names[i] = tool.Name()
	}
	return names
}

func (t *GitHubAgentToolkit) createTools() []tools.Tool {
	var toolList []tools.Tool

	// Core operations
	toolList = append(toolList, &GitHubAgentTool{
		name:        "Get Issues",
		description: "This tool will fetch a list of the repository's issues. It will return the title, and issue number of 5 issues. It takes no input.",
		wrapper:     t.wrapper,
		mode:        "get_issues",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Get Issue",
		description: "This tool will fetch the title, body, and comment thread of a specific issue. **VERY IMPORTANT**: You must specify the issue number as an integer.",
		wrapper:     t.wrapper,
		mode:        "get_issue",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Comment on Issue",
		description: "This tool is useful when you need to comment on a GitHub issue. Simply pass in the issue number and the comment you would like to make. Please use this sparingly as we don't want to clutter the comment threads. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules: - First you must specify the issue number as an integer - Then you must place two newlines - Then you must specify your comment",
		wrapper:     t.wrapper,
		mode:        "comment_on_issue",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "List open pull requests (PRs)",
		description: "This tool will fetch a list of the repository's Pull Requests (PRs). It will return the title, and PR number of 5 PRs. It takes no input.",
		wrapper:     t.wrapper,
		mode:        "list_open_pull_requests",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Get Pull Request",
		description: "This tool will fetch the title, body, comment thread and commit history of a specific Pull Request (by PR number). **VERY IMPORTANT**: You must specify the PR number as an integer.",
		wrapper:     t.wrapper,
		mode:        "get_pull_request",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Create Pull Request",
		description: "This tool is useful when you need to create a new pull request in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules: - First you must specify the title of the pull request - Then you must place two newlines - Then you must write the body or description of the pull request When appropriate, always reference relevant issues in the body by using the syntax `closes #<issue_number` like `closes #3, closes #6`. For example, if you would like to create a pull request called \"README updates\" with contents \"added contributors' names, closes #3\", you would pass in the following string: README updates\n\nadded contributors' names, closes #3",
		wrapper:     t.wrapper,
		mode:        "create_pull_request",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Create File",
		description: "This tool is a wrapper for the GitHub API, useful when you need to create a file in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules: - First you must specify which file to create by passing a full file path (**IMPORTANT**: the path must not start with a slash) - Then you must specify the contents of the file For example, if you would like to create a file called /test/test.txt with contents \"test contents\", you would pass in the following string: test/test.txt\n\ntest contents",
		wrapper:     t.wrapper,
		mode:        "create_file",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Read File",
		description: "This tool is a wrapper for the GitHub API, useful when you need to read the contents of a file. Simply pass in the full file path of the file you would like to read. **IMPORTANT**: the path must not start with a slash",
		wrapper:     t.wrapper,
		mode:        "read_file",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Update File",
		description: "This tool is a wrapper for the GitHub API, useful when you need to update the contents of a file in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules: - First you must specify which file to modify by passing a full file path (**IMPORTANT**: the path must not start with a slash) - Then you must specify the old contents which you would like to replace wrapped in OLD <<<< and >>>> OLD - Then you must specify the new contents which you would like to replace the old contents with wrapped in NEW <<<< and >>>> NEW For example, if you would like to replace the contents of the file /test/test.txt from \"old contents\" to \"new contents\", you would pass in the following string: test/test.txt\nThis is text that will not be changed\nOLD <<<<\nold contents\n>>>> OLD\nNEW <<<<\nnew contents\n>>>> NEW",
		wrapper:     t.wrapper,
		mode:        "update_file",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Delete File",
		description: "This tool is a wrapper for the GitHub API, useful when you need to delete a file in a GitHub repository. Simply pass in the full file path of the file you would like to delete. **IMPORTANT**: the path must not start with a slash",
		wrapper:     t.wrapper,
		mode:        "delete_file",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Overview of existing files in Main branch",
		description: "This tool will provide an overview of all existing files in the main branch of the repository. It will list the file names, their respective paths, and a brief summary of their contents. This can be useful for understanding the structure and content of the repository, especially when navigating through large codebases. No input parameters are required.",
		wrapper:     t.wrapper,
		mode:        "list_files_in_main_branch",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Overview of files in current working branch",
		description: "This tool will provide an overview of all files in your current working branch where you should implement changes. This is great for getting a high level overview of the structure of your code. No input parameters are required.",
		wrapper:     t.wrapper,
		mode:        "list_files_in_bot_branch",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "List branches in this repository",
		description: "This tool will fetch a list of all branches in the repository. It will return the name of each branch. No input parameters are required.",
		wrapper:     t.wrapper,
		mode:        "list_branches_in_repo",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Set active branch",
		description: "This tool will set the active branch in the repository, similar to `git checkout <branch_name>` and `git switch -c <branch_name>`. **VERY IMPORTANT**: You must specify the name of the branch as a string input parameter.",
		wrapper:     t.wrapper,
		mode:        "set_active_branch",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Create a new branch",
		description: "This tool will create a new branch in the repository. **VERY IMPORTANT**: You must specify the name of the new branch as a string input parameter.",
		wrapper:     t.wrapper,
		mode:        "create_branch",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Get files from a directory",
		description: "This tool will fetch a list of all files in a specified directory. **VERY IMPORTANT**: You must specify the path of the directory as a string input parameter.",
		wrapper:     t.wrapper,
		mode:        "get_files_from_directory",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Search issues and pull requests",
		description: "Searches issues and pull requests in the repository. **VERY IMPORTANT**: You must specify the search query as a string input parameter.",
		wrapper:     t.wrapper,
		mode:        "search_issues_and_prs",
	})

	toolList = append(toolList, &GitHubAgentTool{
		name:        "Search code",
		description: "This tool will search for code in the repository. **VERY IMPORTANT**: You must specify the search query as a string input parameter.",
		wrapper:     t.wrapper,
		mode:        "search_code",
	})

	// Optional release tools
	if t.includeReleaseTools {
		toolList = append(toolList, &GitHubAgentTool{
			name:        "Get latest release",
			description: "This tool will fetch the latest release of the repository. No input parameters are required.",
			wrapper:     t.wrapper,
			mode:        "get_latest_release",
		})

		toolList = append(toolList, &GitHubAgentTool{
			name:        "Get releases",
			description: "This tool will fetch the latest 5 releases of the repository. No input parameters are required.",
			wrapper:     t.wrapper,
			mode:        "get_releases",
		})

		toolList = append(toolList, &GitHubAgentTool{
			name:        "Get release",
			description: "This tool will fetch a specific release of the repository. **VERY IMPORTANT**: You must specify the tag name of the release as a string input parameter.",
			wrapper:     t.wrapper,
			mode:        "get_release",
		})
	}

	return toolList
}

// GitHubAgentTool implements the tools.Tool interface for GitHub operations.
type GitHubAgentTool struct {
	name        string
	description string
	wrapper     *githubutil.GitHubAPIWrapper
	mode        string
}

var _ tools.Tool = (*GitHubAgentTool)(nil)

// Name returns the name of the tool.
func (t *GitHubAgentTool) Name() string {
	return t.name
}

// Description returns the description of the tool.
func (t *GitHubAgentTool) Description() string {
	return t.description
}

// Call executes the GitHub operation with the given input.
func (t *GitHubAgentTool) Call(ctx context.Context, input string) (string, error) {
	// Handle input preprocessing based on the mode
	processedInput := t.preprocessInput(input)

	// Use the wrapper's Run method to execute the operation
	result, err := t.wrapper.Run(t.mode, processedInput)
	if err != nil {
		return "", fmt.Errorf("GitHub operation failed: %w", err)
	}

	return result, nil
}

// preprocessInput handles mode-specific input preprocessing.
func (t *GitHubAgentTool) preprocessInput(input string) string {
	input = strings.TrimSpace(input)

	switch t.mode {
	case "get_issue", "get_pull_request":
		// For operations that need integer input, validate and clean
		if num, err := strconv.Atoi(input); err == nil {
			return strconv.Itoa(num)
		}
		// Try to extract number from input
		for _, part := range strings.Fields(input) {
			if num, err := strconv.Atoi(part); err == nil {
				return strconv.Itoa(num)
			}
		}
		return input

	case "comment_on_issue":
		// Ensure proper formatting for comments
		parts := strings.SplitN(input, "\n", 2)
		if len(parts) == 2 {
			// Try to extract issue number and comment
			issueNum := strings.TrimSpace(parts[0])
			comment := strings.TrimSpace(parts[1])
			return fmt.Sprintf("%s\n\n%s", issueNum, comment)
		}
		return input

	case "create_file", "update_file":
		// These modes expect specific formatting, pass through as-is
		return input

	case "create_pull_request":
		// Ensure proper PR formatting
		parts := strings.SplitN(input, "\n", 2)
		if len(parts) == 2 {
			title := strings.TrimSpace(parts[0])
			body := strings.TrimSpace(parts[1])
			return fmt.Sprintf("%s\n\n%s", title, body)
		}
		return input

	default:
		return input
	}
}

package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/tools"
)

// ReadFileTool reads the contents of a file from the repository.
type ReadFileTool struct {
	BaseTool
}

var _ tools.Tool = (*ReadFileTool)(nil)

// NewReadFileTool creates a new tool for reading files.
func NewReadFileTool() (*ReadFileTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &ReadFileTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *ReadFileTool) Name() string {
	return "Read File"
}

// Description returns the description of the tool.
func (t *ReadFileTool) Description() string {
	return "This tool is a wrapper for the GitHub API, useful when you need to read the contents of a file. Simply pass in the full file path of the file you would like to read. **IMPORTANT**: the path must not start with a slash"
}

// Call executes the tool to read a file.
func (t *ReadFileTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	filePath := strings.TrimSpace(input)
	if filePath == "" {
		err := fmt.Errorf("file path cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Remove leading slash if present
	filePath = strings.TrimPrefix(filePath, "/")

	fileContent, _, _, err := t.client.Repositories.GetContents(ctx, t.client.Owner(), t.client.Repo(), filePath, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if fileContent == nil {
		err := fmt.Errorf("file %s not found or is a directory", filePath)
		t.handleToolError(ctx, err)
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to decode file content: %w", err)
	}

	result := fmt.Sprintf("Contents of %s:\n\n%s", filePath, content)
	t.handleToolEnd(ctx, result)
	return result, nil
}

// CreateFileTool creates a new file in the repository.
type CreateFileTool struct {
	BaseTool
}

var _ tools.Tool = (*CreateFileTool)(nil)

// NewCreateFileTool creates a new tool for creating files.
func NewCreateFileTool() (*CreateFileTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &CreateFileTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *CreateFileTool) Name() string {
	return "Create File"
}

// Description returns the description of the tool.
func (t *CreateFileTool) Description() string {
	return `This tool is a wrapper for the GitHub API, useful when you need to create a file in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules:

- First you must specify which file to create by passing a full file path (**IMPORTANT**: the path must not start with a slash)
- Then you must specify the contents of the file

For example, if you would like to create a file called /test/test.txt with contents "test contents", you would pass in the following string:

test/test.txt

test contents`
}

// Call executes the tool to create a file.
func (t *CreateFileTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	parts := strings.SplitN(input, "\n\n", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("invalid input format: expected 'filepath\\n\\ncontents', got: %s", input)
		t.handleToolError(ctx, err)
		return "", err
	}

	filePath := strings.TrimSpace(parts[0])
	content := parts[1] // Don't trim the content as it might be intentionally formatted

	// Remove leading slash if present
	filePath = strings.TrimPrefix(filePath, "/")

	if filePath == "" {
		err := fmt.Errorf("file path cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	message := fmt.Sprintf("Create %s", filePath)
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(content),
	}

	_, _, err := t.client.Repositories.CreateFile(ctx, t.client.Owner(), t.client.Repo(), filePath, opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	result := fmt.Sprintf("Successfully created file: %s", filePath)
	t.handleToolEnd(ctx, result)
	return result, nil
}

// UpdateFileTool updates an existing file in the repository.
type UpdateFileTool struct {
	BaseTool
}

var _ tools.Tool = (*UpdateFileTool)(nil)

// NewUpdateFileTool creates a new tool for updating files.
func NewUpdateFileTool() (*UpdateFileTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &UpdateFileTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *UpdateFileTool) Name() string {
	return "Update File"
}

// Description returns the description of the tool.
func (t *UpdateFileTool) Description() string {
	return `This tool is a wrapper for the GitHub API, useful when you need to update the contents of a file in a GitHub repository. **VERY IMPORTANT**: Your input to this tool MUST strictly follow these rules:

- First you must specify which file to modify by passing a full file path (**IMPORTANT**: the path must not start with a slash)
- Then you must specify the old contents which you would like to replace wrapped in OLD <<<< and >>>> OLD
- Then you must specify the new contents which you would like to replace the old contents with wrapped in NEW <<<< and >>>> NEW

For example, if you would like to replace the contents of the file /test/test.txt from "old contents" to "new contents", you would pass in the following string:

test/test.txt

This is text that will not be changed
OLD <<<<
old contents
>>>> OLD
NEW <<<<
new contents
>>>> NEW`
}

// Call executes the tool to update a file.
func (t *UpdateFileTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	lines := strings.Split(input, "\n")
	if len(lines) < 1 {
		err := fmt.Errorf("invalid input format: missing file path")
		t.handleToolError(ctx, err)
		return "", err
	}

	filePath := strings.TrimSpace(lines[0])
	filePath = strings.TrimPrefix(filePath, "/")

	if filePath == "" {
		err := fmt.Errorf("file path cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Find OLD and NEW sections
	content := strings.Join(lines[1:], "\n")

	oldStart := strings.Index(content, "OLD <<<<")
	oldEnd := strings.Index(content, ">>>> OLD")
	newStart := strings.Index(content, "NEW <<<<")
	newEnd := strings.Index(content, ">>>> NEW")

	if oldStart == -1 || oldEnd == -1 || newStart == -1 || newEnd == -1 {
		err := fmt.Errorf("invalid format: missing OLD <<<< ... >>>> OLD or NEW <<<< ... >>>> NEW markers")
		t.handleToolError(ctx, err)
		return "", err
	}

	oldContent := strings.TrimSpace(content[oldStart+8 : oldEnd])
	newContent := strings.TrimSpace(content[newStart+8 : newEnd])

	// Get current file content and SHA
	fileContent, _, _, err := t.client.Repositories.GetContents(ctx, t.client.Owner(), t.client.Repo(), filePath, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to get current file content for %s: %w", filePath, err)
	}

	if fileContent == nil {
		err := fmt.Errorf("file %s not found", filePath)
		t.handleToolError(ctx, err)
		return "", err
	}

	currentContent, err := fileContent.GetContent()
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to decode current file content: %w", err)
	}

	// Replace old content with new content
	updatedContent := strings.ReplaceAll(currentContent, oldContent, newContent)

	message := fmt.Sprintf("Update %s", filePath)
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(updatedContent),
		SHA:     fileContent.SHA,
	}

	_, _, err = t.client.Repositories.UpdateFile(ctx, t.client.Owner(), t.client.Repo(), filePath, opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to update file %s: %w", filePath, err)
	}

	result := fmt.Sprintf("Successfully updated file: %s", filePath)
	t.handleToolEnd(ctx, result)
	return result, nil
}

// DeleteFileTool deletes a file from the repository.
type DeleteFileTool struct {
	BaseTool
}

var _ tools.Tool = (*DeleteFileTool)(nil)

// NewDeleteFileTool creates a new tool for deleting files.
func NewDeleteFileTool() (*DeleteFileTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &DeleteFileTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *DeleteFileTool) Name() string {
	return "Delete File"
}

// Description returns the description of the tool.
func (t *DeleteFileTool) Description() string {
	return "This tool is a wrapper for the GitHub API, useful when you need to delete a file in a GitHub repository. Simply pass in the full file path of the file you would like to delete. **IMPORTANT**: the path must not start with a slash"
}

// Call executes the tool to delete a file.
func (t *DeleteFileTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	filePath := strings.TrimSpace(input)
	if filePath == "" {
		err := fmt.Errorf("file path cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	// Remove leading slash if present
	filePath = strings.TrimPrefix(filePath, "/")

	// Get current file to get SHA
	fileContent, _, _, err := t.client.Repositories.GetContents(ctx, t.client.Owner(), t.client.Repo(), filePath, nil)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to get file %s for deletion: %w", filePath, err)
	}

	if fileContent == nil {
		err := fmt.Errorf("file %s not found", filePath)
		t.handleToolError(ctx, err)
		return "", err
	}

	message := fmt.Sprintf("Delete %s", filePath)
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		SHA:     fileContent.SHA,
	}

	_, _, err = t.client.Repositories.DeleteFile(ctx, t.client.Owner(), t.client.Repo(), filePath, opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to delete file %s: %w", filePath, err)
	}

	result := fmt.Sprintf("Successfully deleted file: %s", filePath)
	t.handleToolEnd(ctx, result)
	return result, nil
}

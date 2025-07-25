package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/tmc/langchaingo/tools"
)

// GetReleasesTool fetches a list of repository releases.
type GetReleasesTool struct {
	BaseTool
}

var _ tools.Tool = (*GetReleasesTool)(nil)

// NewGetReleasesTool creates a new tool for getting releases.
func NewGetReleasesTool() (*GetReleasesTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetReleasesTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetReleasesTool) Name() string {
	return "Get Releases"
}

// Description returns the description of the tool.
func (t *GetReleasesTool) Description() string {
	return "This tool will fetch the latest 5 releases of the repository. No input parameters are required."
}

// Call executes the tool to get releases.
func (t *GetReleasesTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	opts := &github.ListOptions{
		PerPage: 5,
	}

	releases, _, err := t.client.Repositories.ListReleases(ctx, t.client.Owner(), t.client.Repo(), opts)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}

	var result strings.Builder
	result.WriteString("Repository Releases:\n\n")

	if len(releases) == 0 {
		result.WriteString("No releases found.\n")
	} else {
		for _, release := range releases {
			result.WriteString(fmt.Sprintf("Release: %s\n", release.GetTagName()))
			result.WriteString(fmt.Sprintf("Name: %s\n", release.GetName()))
			result.WriteString(fmt.Sprintf("Published: %s\n", release.GetPublishedAt().Format("2006-01-02 15:04:05")))
			result.WriteString(fmt.Sprintf("Draft: %t, Prerelease: %t\n", release.GetDraft(), release.GetPrerelease()))
			if release.GetBody() != "" {
				result.WriteString(fmt.Sprintf("Description: %s\n", release.GetBody()))
			}
			result.WriteString("\n---\n\n")
		}
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// GetLatestReleaseTool fetches the latest release.
type GetLatestReleaseTool struct {
	BaseTool
}

var _ tools.Tool = (*GetLatestReleaseTool)(nil)

// NewGetLatestReleaseTool creates a new tool for getting the latest release.
func NewGetLatestReleaseTool() (*GetLatestReleaseTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetLatestReleaseTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetLatestReleaseTool) Name() string {
	return "Get Latest Release"
}

// Description returns the description of the tool.
func (t *GetLatestReleaseTool) Description() string {
	return "This tool will fetch the latest release of the repository. No input parameters are required."
}

// Call executes the tool to get the latest release.
func (t *GetLatestReleaseTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	release, _, err := t.client.Repositories.GetLatestRelease(ctx, t.client.Owner(), t.client.Repo())
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}

	var result strings.Builder
	result.WriteString("Latest Release:\n\n")
	result.WriteString(fmt.Sprintf("Tag: %s\n", release.GetTagName()))
	result.WriteString(fmt.Sprintf("Name: %s\n", release.GetName()))
	result.WriteString(fmt.Sprintf("Published: %s\n", release.GetPublishedAt().Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Draft: %t, Prerelease: %t\n", release.GetDraft(), release.GetPrerelease()))

	if release.GetBody() != "" {
		result.WriteString(fmt.Sprintf("Description:\n%s\n", release.GetBody()))
	}

	if len(release.Assets) > 0 {
		result.WriteString("\nAssets:\n")
		for _, asset := range release.Assets {
			result.WriteString(fmt.Sprintf("- %s (%d bytes)\n", asset.GetName(), asset.GetSize()))
		}
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

// GetReleaseTool fetches a specific release by tag name.
type GetReleaseTool struct {
	BaseTool
}

var _ tools.Tool = (*GetReleaseTool)(nil)

// NewGetReleaseTool creates a new tool for getting a specific release.
func NewGetReleaseTool() (*GetReleaseTool, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &GetReleaseTool{
		BaseTool: BaseTool{client: client},
	}, nil
}

// Name returns the name of the tool.
func (t *GetReleaseTool) Name() string {
	return "Get Release"
}

// Description returns the description of the tool.
func (t *GetReleaseTool) Description() string {
	return "This tool will fetch a specific release of the repository. **VERY IMPORTANT**: You must specify the tag name of the release as a string input parameter."
}

// Call executes the tool to get a specific release.
func (t *GetReleaseTool) Call(ctx context.Context, input string) (string, error) {
	t.handleToolStart(ctx, input)

	tagName := strings.TrimSpace(input)
	if tagName == "" {
		err := fmt.Errorf("tag name cannot be empty")
		t.handleToolError(ctx, err)
		return "", err
	}

	release, _, err := t.client.Repositories.GetReleaseByTag(ctx, t.client.Owner(), t.client.Repo(), tagName)
	if err != nil {
		t.handleToolError(ctx, err)
		return "", fmt.Errorf("failed to fetch release %s: %w", tagName, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Release %s:\n\n", tagName))
	result.WriteString(fmt.Sprintf("Tag: %s\n", release.GetTagName()))
	result.WriteString(fmt.Sprintf("Name: %s\n", release.GetName()))
	result.WriteString(fmt.Sprintf("Published: %s\n", release.GetPublishedAt().Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Draft: %t, Prerelease: %t\n", release.GetDraft(), release.GetPrerelease()))

	if release.GetBody() != "" {
		result.WriteString(fmt.Sprintf("Description:\n%s\n", release.GetBody()))
	}

	if len(release.Assets) > 0 {
		result.WriteString("\nAssets:\n")
		for _, asset := range release.Assets {
			result.WriteString(fmt.Sprintf("- %s (%d bytes)\n", asset.GetName(), asset.GetSize()))
		}
	}

	output := result.String()
	t.handleToolEnd(ctx, output)
	return output, nil
}

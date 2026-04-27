package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
)

// getGitHubClient creates a GitHub client using GITHUB_TOKEN env var.
func getGitHubClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf(
			"GITHUB_TOKEN environment variable is not set. " +
				"Create a PAT at https://github.com/settings/tokens (scope: repo) " +
				"and add it to your shell profile or environment",
		)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	baseURL := os.Getenv("GITHUB_API_URL")
	if baseURL != "" {
		client, err := github.NewClient(tc).WithAuthToken(token).WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub enterprise client: %w", err)
		}
		return client, nil
	}

	return github.NewClient(tc), nil
}

// logRateLimit writes rate limit info to stderr.
func logRateLimit(resp *http.Response) {
	if resp != nil {
		fmt.Fprintf(os.Stderr, "[github-reader] rate-limit remaining: %s, reset: %s\n",
			resp.Header.Get("X-RateLimit-Remaining"), resp.Header.Get("X-RateLimit-Reset"))
	}
}

const (
	defaultPageSize = 8000
	maxPageSize     = 32000
)

// RegisterGitHubTool registers the GitHub MCP tools.
func RegisterGitHubTool(s *server.MCPServer) {
	// Tool: get_file_content
	s.AddTool(
		mcp.NewTool("github",
			mcp.WithDescription("GitHub operations. Use 'action' to specify: get_file_content, search_repos, list_commits."),
			mcp.WithString("action",
				mcp.Required(),
				mcp.Description("Action to perform"),
				mcp.Enum("get_file_content", "search_repos", "list_commits"),
			),
			mcp.WithString("owner", mcp.Description("Repository owner (required for get_file_content, list_commits)")),
			mcp.WithString("repo", mcp.Description("Repository name (required for get_file_content, list_commits)")),
			mcp.WithString("path", mcp.Description("File path in repository (required for get_file_content)")),
			mcp.WithString("ref", mcp.Description("Git reference (branch, tag, commit) for get_file_content")),
			mcp.WithNumber("page", mcp.Description("Page number for pagination (default: 1)")),
			mcp.WithNumber("page_size", mcp.Description("Characters per page (default: 8000, max: 32000)")),
			mcp.WithString("query", mcp.Description("Search query for search_repos")),
			mcp.WithNumber("per_page", mcp.Description("Results per page for search (default: 10)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return errResult("action is required")
			}
			switch action {
			case "get_file_content":
				return handleGetFileContent(req)
			case "search_repos":
				return handleSearchRepos(req)
			case "list_commits":
				return handleListCommits(req)
			default:
				return errResultf("unknown github action: %s", action)
			}
		},
	)
}

func handleGetFileContent(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	path, _ := args["path"].(string)
	ref, _ := args["ref"].(string)

	if owner == "" || repo == "" || path == "" {
		return errResult("owner, repo, and path are required")
	}

	page := 1
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := defaultPageSize
	if v, ok := args["page_size"].(float64); ok && v > 0 {
		pageSize = int(v)
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
	}

	client, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	opt := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opt.Ref = ref
	}

	fileContent, dirContent, _, err := client.Repositories.GetContents(context.Background(), owner, repo, path, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get content: %v", err)), nil
	}

	if fileContent != nil {
		content, err := fileContent.GetContent()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to decode content: %v", err)), nil
		}

		// Handle pagination
		totalLen := len(content)
		start := (page - 1) * pageSize
		if start >= totalLen {
			return mcp.NewToolResultText(fmt.Sprintf("End of content (page %d, %d/%d chars)", page, totalLen, totalLen)), nil
		}
		end := start + pageSize
		if end > totalLen {
			end = totalLen
		}

		truncated := content[start:end]
		info := fmt.Sprintf("\n--- File: %s/%s/%s (page %d, showing %d-%d of %d chars) ---\n",
			owner, repo, path, page, start+1, end, totalLen)
		if page > 1 {
			info = fmt.Sprintf("\n--- (continued from page %d) ---\n", page-1) + info
		}

		hasMore := end < totalLen
		if hasMore {
			info += fmt.Sprintf("\n[MORE: next page at page %d]", page+1)
		}

		return mcp.NewToolResultText(info + truncated), nil
	}

	if dirContent != nil {
		var entries []string
		for _, entry := range dirContent {
			entries = append(entries, fmt.Sprintf("- %s (%s)", *entry.Name, *entry.Type))
		}
		return mcp.NewToolResultText(fmt.Sprintf("Directory contents of %s/%s/%s:\n%s", owner, repo, path, strings.Join(entries, "\n"))), nil
	}

	return mcp.NewToolResultText("Empty or unknown content type"), nil
}

func handleSearchRepos(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, err := req.RequireString("query")
	if err != nil {
		return errResult("query is required")
	}

	perPage := 10
	if v, ok := args["per_page"].(float64); ok && v > 0 {
		perPage = int(v)
	}

	client, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	results, resp, err := client.Search.Repositories(context.Background(), query, &github.SearchOptions{ListOptions: github.ListOptions{PerPage: perPage}})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}
	logRateLimit(resp.Response)

	if len(results.Repositories) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No repositories found for: %s", query)), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("🔍 Search results for '%s' (%d total):\n\n", query, results.GetTotal()))

	for i, repo := range results.Repositories {
		output.WriteString(fmt.Sprintf("%d. [%s/%s](%s)\n", i+1, *repo.Owner.Login, *repo.Name, *repo.HTMLURL))
		output.WriteString(fmt.Sprintf("   Description: %s\n", stringOrNone(repo.Description)))
		output.WriteString(fmt.Sprintf("   Stars: %s | Language: %s | Updated: %s\n\n",
			strconv.Itoa(repo.GetStargazersCount()),
			stringOrNone(repo.Language),
			repo.GetUpdatedAt().Format("2006-01-02")))
	}

	return mcp.NewToolResultText(output.String()), nil
}

func handleListCommits(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	owner, err := req.RequireString("owner")
	if err != nil {
		return errResult("owner is required")
	}
	repo, err := req.RequireString("repo")
	if err != nil {
		return errResult("repo is required")
	}

	client, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	perPage := 10
	if v, ok := args["per_page"].(float64); ok && v > 0 {
		perPage = int(v)
	}

	commits, resp, err := client.Repositories.ListCommits(context.Background(), owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list commits: %v", err)), nil
	}
	logRateLimit(resp.Response)

	if len(commits) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No commits found for %s/%s", owner, repo)), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📜 Recent commits for [%s/%s](https://github.com/%s/%s/commits):\n\n", owner, repo, owner, repo))

	for i, commit := range commits {
		msg := "No message"
		if commit.Commit != nil && commit.Commit.Message != nil {
			msg = *commit.Commit.Message
			if idx := strings.Index(msg, "\n"); idx != -1 {
				msg = msg[:idx]
			}
		}
		sha := "unknown"
		if commit.SHA != nil {
			sha = (*commit.SHA)[:7]
		}
		date := "unknown"
		if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Author.Date != nil {
			date = commit.Commit.Author.Date.Format("2006-01-02 15:04")
		}
		output.WriteString(fmt.Sprintf("%d. [%s](https://github.com/%s/%s/commit/%s) - %s (%s)\n",
			i+1, sha, owner, repo, sha, msg, date))
	}

	return mcp.NewToolResultText(output.String()), nil
}

func stringOrNone(s *string) string {
	if s == nil || *s == "" {
		return "N/A"
	}
	return *s
}
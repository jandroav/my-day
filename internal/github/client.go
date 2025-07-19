package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default GitHub API base URL
	DefaultBaseURL = "https://api.github.com"
	
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client represents a GitHub API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new GitHub client with token authentication
func NewClient(token string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		token:      token,
	}
}

// NewClientWithURL creates a new GitHub client with custom base URL (for GitHub Enterprise)
func NewClientWithURL(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: DefaultTimeout},
		token:      token,
	}
}

// makeRequest makes an authenticated HTTP request to the GitHub API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil && len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "my-day-cli/1.0")

	return c.httpClient.Do(req)
}

// GetCurrentUser returns information about the authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.makeRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// GetUserRepositories returns repositories for the authenticated user
func (c *Client) GetUserRepositories(ctx context.Context, since time.Time) ([]Repository, error) {
	params := url.Values{
		"affiliation": {"owner,collaborator"},
		"sort":        {"updated"},
		"direction":   {"desc"},
		"per_page":    {"100"},
	}

	if !since.IsZero() {
		params.Set("since", since.Format(time.RFC3339))
	}

	resp, err := c.makeRequest(ctx, "GET", "/user/repos", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repositories response: %w", err)
	}

	return repos, nil
}

// GetPullRequests returns pull requests for a repository
func (c *Client) GetPullRequests(ctx context.Context, owner, repo string, state string, since time.Time) ([]PullRequest, error) {
	params := url.Values{
		"state":     {state}, // open, closed, all
		"sort":      {"updated"},
		"direction": {"desc"},
		"per_page":  {"100"},
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var pullRequests []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pullRequests); err != nil {
		return nil, fmt.Errorf("failed to decode pull requests response: %w", err)
	}

	// Filter by since time if provided
	if !since.IsZero() {
		var filtered []PullRequest
		for _, pr := range pullRequests {
			if pr.UpdatedAt.Time.After(since) {
				filtered = append(filtered, pr)
			}
		}
		pullRequests = filtered
	}

	return pullRequests, nil
}

// GetUserPullRequests returns pull requests involving the authenticated user
func (c *Client) GetUserPullRequests(ctx context.Context, repos []Repository, since time.Time) ([]PullRequest, error) {
	var allPRs []PullRequest

	user, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	for _, repo := range repos {
		// Get PRs where user is author or reviewer
		prs, err := c.GetPullRequests(ctx, repo.Owner.Login, repo.Name, "all", since)
		if err != nil {
			// Log error but continue with other repos
			continue
		}

		// Filter PRs involving the current user
		for _, pr := range prs {
			if c.isUserInvolvedInPR(pr, user.Login) {
				pr.Repository = repo // Ensure repository info is attached
				allPRs = append(allPRs, pr)
			}
		}
	}

	return allPRs, nil
}

// isUserInvolvedInPR checks if a user is involved in a pull request
func (c *Client) isUserInvolvedInPR(pr PullRequest, username string) bool {
	// Author
	if pr.User.Login == username {
		return true
	}

	// Assignees
	for _, assignee := range pr.Assignees {
		if assignee.Login == username {
			return true
		}
	}

	// Reviewers
	for _, reviewer := range pr.Reviewers {
		if reviewer.Login == username {
			return true
		}
	}

	return false
}

// GetCommits returns commits for a repository
func (c *Client) GetCommits(ctx context.Context, owner, repo string, since time.Time, author string) ([]Commit, error) {
	params := url.Values{
		"per_page": {"100"},
	}

	if !since.IsZero() {
		params.Set("since", since.Format(time.RFC3339))
	}

	if author != "" {
		params.Set("author", author)
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/commits", owner, repo)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("failed to decode commits response: %w", err)
	}

	return commits, nil
}

// GetWorkflowRuns returns workflow runs for a repository
func (c *Client) GetWorkflowRuns(ctx context.Context, owner, repo string, since time.Time) ([]WorkflowRun, error) {
	params := url.Values{
		"per_page": {"100"},
	}

	if !since.IsZero() {
		params.Set("created", ">"+since.Format("2006-01-02"))
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/actions/runs", owner, repo)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow runs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var response struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode workflow runs response: %w", err)
	}

	return response.WorkflowRuns, nil
}

// GetUserActivity returns unified activity for the authenticated user
func (c *Client) GetUserActivity(ctx context.Context, since time.Time, repos []string) ([]Activity, error) {
	var activities []Activity

	// Get user info
	user, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Get repositories to analyze
	var repositories []Repository
	if len(repos) > 0 {
		// Specific repositories requested
		for _, repoName := range repos {
			parts := strings.Split(repoName, "/")
			if len(parts) != 2 {
				continue
			}
			// We'll need to implement GetRepository if we want to validate specific repos
		}
	} else {
		// Get all user repositories
		repositories, err = c.GetUserRepositories(ctx, since)
		if err != nil {
			return nil, fmt.Errorf("failed to get repositories: %w", err)
		}
	}

	// Get activity from each repository
	for _, repo := range repositories {
		repoActivities, err := c.getRepositoryActivity(ctx, repo, user.Login, since)
		if err != nil {
			// Log error but continue with other repos
			continue
		}
		activities = append(activities, repoActivities...)
	}

	return activities, nil
}

// getRepositoryActivity gets all activity for a specific repository
func (c *Client) getRepositoryActivity(ctx context.Context, repo Repository, username string, since time.Time) ([]Activity, error) {
	var activities []Activity

	// Get commits
	commits, err := c.GetCommits(ctx, repo.Owner.Login, repo.Name, since, username)
	if err == nil {
		for _, commit := range commits {
			activity := Activity{
				Type:        "commit",
				ID:          commit.SHA,
				Title:       commit.Message,
				Description: fmt.Sprintf("Commit to %s", repo.FullName),
				State:       "committed",
				URL:         commit.HTMLURL,
				Repository:  repo.FullName,
				Author:      username,
				CreatedAt:   commit.Author.Date.Time,
				UpdatedAt:   commit.Author.Date.Time,
				JiraTickets: extractJiraTickets(commit.Message),
				Metadata: map[string]interface{}{
					"sha":     commit.SHA,
					"tree":    commit.Tree.SHA,
					"parents": commit.Parents,
				},
			}
			activities = append(activities, activity)
		}
	}

	// Get pull requests
	prs, err := c.GetPullRequests(ctx, repo.Owner.Login, repo.Name, "all", since)
	if err == nil {
		for _, pr := range prs {
			if c.isUserInvolvedInPR(pr, username) {
				state := pr.State
				if pr.Merged {
					state = "merged"
				}

				activity := Activity{
					Type:        "pull_request",
					ID:          strconv.FormatInt(pr.ID, 10),
					Title:       pr.Title,
					Description: pr.Body,
					State:       state,
					URL:         pr.HTMLURL,
					Repository:  repo.FullName,
					Author:      pr.User.Login,
					CreatedAt:   pr.CreatedAt.Time,
					UpdatedAt:   pr.UpdatedAt.Time,
					JiraTickets: extractJiraTickets(pr.Title + " " + pr.Body),
					Metadata: map[string]interface{}{
						"number":    pr.Number,
						"draft":     pr.Draft,
						"merged":    pr.Merged,
						"head_ref":  pr.Head.Ref,
						"base_ref":  pr.Base.Ref,
						"mergeable": pr.State == "open",
					},
				}
				activities = append(activities, activity)
			}
		}
	}

	return activities, nil
}

// extractJiraTickets extracts Jira ticket references from text
func extractJiraTickets(text string) []string {
	// Common Jira ticket patterns: PROJECT-123, ABC-456, etc.
	re := regexp.MustCompile(`\b([A-Z]{2,10}-\d+)\b`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	var tickets []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			ticket := match[1]
			if !seen[ticket] {
				tickets = append(tickets, ticket)
				seen[ticket] = true
			}
		}
	}
	
	return tickets
}

// TestConnection tests the GitHub API connection
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("GitHub connection test failed: %w", err)
	}
	return nil
}
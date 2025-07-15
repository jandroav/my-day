package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Client represents a Jira API client
type Client struct {
	baseURL     string
	cloudID     string
	httpClient  *http.Client
	authManager *AuthManager
}

// NewClient creates a new Jira client
func NewClient(baseURL, clientID, clientSecret, redirectURI string) *Client {
	authManager := NewAuthManager(clientID, clientSecret, redirectURI)
	
	return &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		authManager: authManager,
	}
}

// GetAuthManager returns the authentication manager
func (c *Client) GetAuthManager() *AuthManager {
	return c.authManager
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, maxResults int) (*SearchResponse, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Get cloud ID if not already set
	if c.cloudID == "" {
		cloudID, err := c.getCloudID(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get cloud ID: %w", err)
		}
		c.cloudID = cloudID
	}

	// Build search URL using the proper cloud-specific endpoint
	searchURL := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/search", c.cloudID)
	params := url.Values{
		"jql":        {jql},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
		"fields":     {"summary,description,status,priority,issuetype,project,assignee,reporter,created,updated,resolution,labels"},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResponse, nil
}

// GetIssuesByProjects retrieves issues for specific projects
func (c *Client) GetIssuesByProjects(ctx context.Context, projectKeys []string, maxResults int) (*SearchResponse, error) {
	if len(projectKeys) == 0 {
		return &SearchResponse{Issues: []Issue{}}, nil
	}

	// Build JQL query for project keys
	projectFilter := strings.Join(projectKeys, ",")
	jql := fmt.Sprintf("project in (%s) ORDER BY updated DESC", projectFilter)

	return c.SearchIssues(ctx, jql, maxResults)
}

// GetMyRecentIssues retrieves issues assigned to or reported by the current user
func (c *Client) GetMyRecentIssues(ctx context.Context, projectKeys []string, maxResults int) (*SearchResponse, error) {
	var jqlParts []string

	// Add project filter if specified
	if len(projectKeys) > 0 {
		projectFilter := strings.Join(projectKeys, ",")
		jqlParts = append(jqlParts, fmt.Sprintf("project in (%s)", projectFilter))
	}

	// Add user filter (assignee or reporter)
	jqlParts = append(jqlParts, "(assignee = currentUser() OR reporter = currentUser())")

	// Add time filter for recent activity
	jqlParts = append(jqlParts, "updated >= -7d")

	jql := strings.Join(jqlParts, " AND ") + " ORDER BY updated DESC"

	return c.SearchIssues(ctx, jql, maxResults)
}

// GetMyWorklog retrieves worklog entries for the current user
func (c *Client) GetMyWorklog(ctx context.Context, since time.Time) ([]WorklogEntry, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Get current user info first
	userInfo, err := c.getCurrentUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Search for issues with worklog by current user
	sinceStr := since.Format("2006-01-02")
	jql := fmt.Sprintf("worklogAuthor = currentUser() AND worklogDate >= %s ORDER BY updated DESC", sinceStr)

	searchResponse, err := c.SearchIssues(ctx, jql, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to search worklog issues: %w", err)
	}

	var worklogEntries []WorklogEntry
	for _, issue := range searchResponse.Issues {
		worklogs, err := c.getIssueWorklogs(ctx, token, issue.Key, userInfo.AccountID, since)
		if err != nil {
			continue // Skip issues where we can't get worklogs
		}
		worklogEntries = append(worklogEntries, worklogs...)
	}

	return worklogEntries, nil
}

// GetCurrentUser gets information about the current authenticated user (public method)
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}
	
	return c.getCurrentUser(ctx, token)
}

// getCurrentUser gets information about the current authenticated user
func (c *Client) getCurrentUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	// Get cloud ID if not already set
	if c.cloudID == "" {
		cloudID, err := c.getCloudID(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get cloud ID: %w", err)
		}
		c.cloudID = cloudID
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/myself", c.cloudID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetMyIssuesWithTodaysComments retrieves issues where the current user added comments today
func (c *Client) GetMyIssuesWithTodaysComments(ctx context.Context, projectKeys []string, maxResults int) (*SearchResponse, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Get current user info first
	userInfo, err := c.getCurrentUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var jqlParts []string

	// Add project filter if specified
	if len(projectKeys) > 0 {
		projectFilter := strings.Join(projectKeys, ",")
		jqlParts = append(jqlParts, fmt.Sprintf("project in (%s)", projectFilter))
	}

	// Add user filter (assignee or reporter) - get issues user is involved with
	jqlParts = append(jqlParts, "(assignee = currentUser() OR reporter = currentUser())")

	// Get recent issues first, then filter by comments client-side
	jqlParts = append(jqlParts, "updated >= -7d")

	jql := strings.Join(jqlParts, " AND ") + " ORDER BY updated DESC"

	searchResponse, err := c.SearchIssues(ctx, jql, maxResults*2) // Get more to filter down
	if err != nil {
		return nil, err
	}

	// Filter issues by today's comments
	var filteredIssues []Issue
	today := time.Now().Truncate(24 * time.Hour)

	for _, issue := range searchResponse.Issues {
		hasCommentToday, err := c.hasUserCommentToday(ctx, token, issue.Key, userInfo.AccountID, today)
		if err != nil {
			continue // Skip issues where we can't check comments
		}
		if hasCommentToday {
			filteredIssues = append(filteredIssues, issue)
			if len(filteredIssues) >= maxResults {
				break
			}
		}
	}

	return &SearchResponse{
		Issues:     filteredIssues,
		Total:      len(filteredIssues),
		MaxResults: maxResults,
		StartAt:    0,
	}, nil
}

// hasUserCommentToday checks if the user added a comment to the issue today
func (c *Client) hasUserCommentToday(ctx context.Context, token *oauth2.Token, issueKey, userAccountID string, today time.Time) (bool, error) {
	// Get cloud ID if not already set
	if c.cloudID == "" {
		cloudID, err := c.getCloudID(ctx, token)
		if err != nil {
			return false, fmt.Errorf("failed to get cloud ID: %w", err)
		}
		c.cloudID = cloudID
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/issue/%s/comment", c.cloudID, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get comments: status %d", resp.StatusCode)
	}

	var commentResponse struct {
		Comments []Comment `json:"comments"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&commentResponse); err != nil {
		return false, err
	}

	// Check if user has any comments today
	for _, comment := range commentResponse.Comments {
		if comment.Author.AccountID == userAccountID {
			commentDate := comment.Created.Time.Truncate(24 * time.Hour)
			if commentDate.Equal(today) {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetUserCommentsToday retrieves user's comments from today for a specific issue
func (c *Client) GetUserCommentsToday(ctx context.Context, issueKey, userAccountID string, today time.Time) ([]Comment, error) {
	// Get cloud ID if not already set
	if c.cloudID == "" {
		token, err := c.authManager.GetValidToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("authentication required: %w", err)
		}
		
		cloudID, err := c.getCloudID(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get cloud ID: %w", err)
		}
		c.cloudID = cloudID
	}

	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/issue/%s/comment", c.cloudID, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get comments: status %d", resp.StatusCode)
	}

	var commentResponse struct {
		Comments []Comment `json:"comments"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&commentResponse); err != nil {
		return nil, err
	}

	// Filter for user's comments from today
	var todaysComments []Comment
	for _, comment := range commentResponse.Comments {
		if comment.Author.AccountID == userAccountID {
			commentDate := comment.Created.Time.Truncate(24 * time.Hour)
			if commentDate.Equal(today) {
				todaysComments = append(todaysComments, comment)
			}
		}
	}

	return todaysComments, nil
}

// getIssueWorklogs retrieves worklog entries for a specific issue
func (c *Client) getIssueWorklogs(ctx context.Context, token *oauth2.Token, issueKey, userAccountID string, since time.Time) ([]WorklogEntry, error) {
	// Get cloud ID if not already set
	if c.cloudID == "" {
		cloudID, err := c.getCloudID(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get cloud ID: %w", err)
		}
		c.cloudID = cloudID
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/issue/%s/worklog", c.cloudID, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get worklogs: status %d", resp.StatusCode)
	}

	var worklogResponse struct {
		Worklogs []WorklogEntry `json:"worklogs"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&worklogResponse); err != nil {
		return nil, err
	}

	// Filter worklogs by user and date
	var userWorklogs []WorklogEntry
	for _, worklog := range worklogResponse.Worklogs {
		if worklog.Author.AccountID == userAccountID && worklog.Started.Time.After(since) {
			worklog.IssueID = issueKey
			userWorklogs = append(userWorklogs, worklog)
		}
	}

	return userWorklogs, nil
}

// TestConnection tests the connection to Jira using the proper OAuth 2.0 flow
func (c *Client) TestConnection(ctx context.Context) error {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	// First, get accessible resources to find the correct cloud ID
	cloudID, err := c.getCloudID(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get cloud ID: %w", err)
	}

	c.cloudID = cloudID

	// Now test connection using the proper cloud-specific URL
	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/myself", cloudID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	return nil
}

// getCloudID retrieves the cloud ID for the accessible Jira instance
func (c *Client) getCloudID(ctx context.Context, token *oauth2.Token) (string, error) {
	url := "https://api.atlassian.com/oauth/token/accessible-resources"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("accessible-resources API returned status %d", resp.StatusCode)
	}

	var resources []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return "", err
	}

	if len(resources) == 0 {
		return "", fmt.Errorf("no accessible Jira resources found")
	}

	// Find the resource that matches our base URL
	for _, resource := range resources {
		if strings.Contains(resource.URL, strings.TrimPrefix(c.baseURL, "https://")) {
			return resource.ID, nil
		}
	}

	// If no exact match, return the first resource
	return resources[0].ID, nil
}
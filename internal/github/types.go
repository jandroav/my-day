package github

import (
	"encoding/json"
	"strings"
	"time"
)

// GitHubTime represents a time field that can handle GitHub's RFC3339 format
type GitHubTime struct {
	time.Time
}

// UnmarshalJSON handles GitHub's date format
func (gt *GitHubTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	timeStr := strings.Trim(string(data), `"`)
	
	if timeStr == "null" || timeStr == "" {
		gt.Time = time.Time{}
		return nil
	}
	
	// GitHub uses RFC3339 format
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		gt.Time = t
		return nil
	}
	
	// Fallback to RFC3339Nano
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		gt.Time = t
		return nil
	}
	
	return &time.ParseError{
		Layout: "GitHub time format (RFC3339)",
		Value:  timeStr,
	}
}

// MarshalJSON converts GitHubTime back to JSON
func (gt GitHubTime) MarshalJSON() ([]byte, error) {
	if gt.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(gt.Time.Format(time.RFC3339))
}

// Repository represents a GitHub repository
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Language    string `json:"language"`
	Owner       User   `json:"owner"`
	CreatedAt   GitHubTime `json:"created_at"`
	UpdatedAt   GitHubTime `json:"updated_at"`
	PushedAt    GitHubTime `json:"pushed_at"`
}

// User represents a GitHub user
type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	State       string     `json:"state"`       // open, closed
	Merged      bool       `json:"merged"`
	Draft       bool       `json:"draft"`
	HTMLURL     string     `json:"html_url"`
	User        User       `json:"user"`
	Assignees   []User     `json:"assignees"`
	Reviewers   []User     `json:"requested_reviewers"`
	CreatedAt   GitHubTime `json:"created_at"`
	UpdatedAt   GitHubTime `json:"updated_at"`
	ClosedAt    *GitHubTime `json:"closed_at"`
	MergedAt    *GitHubTime `json:"merged_at"`
	Head        Branch     `json:"head"`
	Base        Branch     `json:"base"`
	Repository  Repository `json:"repository"`
	Links       PRLinks    `json:"_links"`
}

// Branch represents a git branch reference
type Branch struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	SHA   string     `json:"sha"`
	User  User       `json:"user"`
	Repo  Repository `json:"repo"`
}

// PRLinks represents links in a pull request
type PRLinks struct {
	Self           Link `json:"self"`
	HTML           Link `json:"html"`
	Issue          Link `json:"issue"`
	Comments       Link `json:"comments"`
	ReviewComments Link `json:"review_comments"`
	Commits        Link `json:"commits"`
}

// Link represents a URL link
type Link struct {
	HREF string `json:"href"`
}

// Commit represents a GitHub commit
type Commit struct {
	SHA       string     `json:"sha"`
	Message   string     `json:"message"`
	Author    CommitUser `json:"author"`
	Committer CommitUser `json:"committer"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
	Tree      Tree       `json:"tree"`
	Parents   []Tree     `json:"parents"`
}

// CommitUser represents the author/committer of a commit
type CommitUser struct {
	Name  string     `json:"name"`
	Email string     `json:"email"`
	Date  GitHubTime `json:"date"`
}

// Tree represents a git tree object
type Tree struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	State       string     `json:"state"`       // open, closed
	Locked      bool       `json:"locked"`
	HTMLURL     string     `json:"html_url"`
	User        User       `json:"user"`
	Assignees   []User     `json:"assignees"`
	Labels      []Label    `json:"labels"`
	Milestone   *Milestone `json:"milestone"`
	CreatedAt   GitHubTime `json:"created_at"`
	UpdatedAt   GitHubTime `json:"updated_at"`
	ClosedAt    *GitHubTime `json:"closed_at"`
	Repository  Repository `json:"repository"`
}

// Label represents a GitHub label
type Label struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Milestone represents a GitHub milestone
type Milestone struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	CreatedAt   GitHubTime `json:"created_at"`
	UpdatedAt   GitHubTime `json:"updated_at"`
	DueOn       *GitHubTime `json:"due_on"`
	ClosedAt    *GitHubTime `json:"closed_at"`
}

// Review represents a pull request review
type Review struct {
	ID             int64      `json:"id"`
	User           User       `json:"user"`
	Body           string     `json:"body"`
	State          string     `json:"state"`          // APPROVED, CHANGES_REQUESTED, COMMENTED
	HTMLURL        string     `json:"html_url"`
	PullRequestURL string     `json:"pull_request_url"`
	SubmittedAt    GitHubTime `json:"submitted_at"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	DisplayTitle string     `json:"display_title"`
	Status       string     `json:"status"`       // queued, in_progress, completed
	Conclusion   string     `json:"conclusion"`   // success, failure, cancelled, etc.
	WorkflowID   int64      `json:"workflow_id"`
	URL          string     `json:"url"`
	HTMLURL      string     `json:"html_url"`
	CreatedAt    GitHubTime `json:"created_at"`
	UpdatedAt    GitHubTime `json:"updated_at"`
	RunStartedAt GitHubTime `json:"run_started_at"`
	HeadBranch   string     `json:"head_branch"`
	HeadSHA      string     `json:"head_sha"`
	Event        string     `json:"event"`
	Actor        User       `json:"actor"`
	Repository   Repository `json:"repository"`
}

// Release represents a GitHub release
type Release struct {
	ID          int64      `json:"id"`
	TagName     string     `json:"tag_name"`
	Name        string     `json:"name"`
	Body        string     `json:"body"`
	Draft       bool       `json:"draft"`
	Prerelease  bool       `json:"prerelease"`
	CreatedAt   GitHubTime `json:"created_at"`
	PublishedAt GitHubTime `json:"published_at"`
	Author      User       `json:"author"`
	HTMLURL     string     `json:"html_url"`
	Repository  Repository `json:"repository"`
}

// AuthInfo represents stored GitHub authentication information
type AuthInfo struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Activity represents a unified activity item from GitHub
type Activity struct {
	Type        string     `json:"type"`         // pr, commit, issue, review, workflow, release
	ID          string     `json:"id"`           // Unique identifier
	Title       string     `json:"title"`        // Human readable title
	Description string     `json:"description"`  // Additional details
	State       string     `json:"state"`        // Current state
	URL         string     `json:"url"`          // Link to GitHub
	Repository  string     `json:"repository"`   // Repository full name
	Author      string     `json:"author"`       // Author username
	CreatedAt   time.Time  `json:"created_at"`   // When it was created
	UpdatedAt   time.Time  `json:"updated_at"`   // When it was last updated
	JiraTickets []string   `json:"jira_tickets"` // Linked Jira ticket keys
	Metadata    map[string]interface{} `json:"metadata"` // Additional type-specific data
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	TotalCount int         `json:"total_count,omitempty"`
	Items      interface{} `json:"items,omitempty"`
}

// ErrorResponse represents a GitHub API error response
type ErrorResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return e.Message
}
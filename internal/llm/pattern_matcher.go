package llm

import (
	"regexp"
	"strings"
	"time"
)

// TechnicalPatternMatcher handles pattern matching for DevOps and technical activities
type TechnicalPatternMatcher struct {
	infrastructurePatterns map[string]*PatternDefinition
	deploymentPatterns     map[string]*PatternDefinition
	developmentPatterns    map[string]*PatternDefinition
	databasePatterns       map[string]*PatternDefinition
	securityPatterns       map[string]*PatternDefinition
	testingPatterns        map[string]*PatternDefinition
	debug                  bool
}

// PatternDefinition defines a pattern with its matching criteria and confidence scoring
type PatternDefinition struct {
	Name        string            `json:"name"`
	Keywords    []string          `json:"keywords"`
	Regex       *regexp.Regexp    `json:"-"`
	Category    string            `json:"category"`
	Subcategory string            `json:"subcategory"`
	BaseScore   float64           `json:"base_score"`
	Modifiers   map[string]float64 `json:"modifiers"`
	Examples    []string          `json:"examples"`
}

// PatternMatch represents a matched pattern with confidence score
type PatternMatch struct {
	Pattern     *PatternDefinition `json:"pattern"`
	Text        string             `json:"text"`
	Confidence  float64            `json:"confidence"`
	Context     string             `json:"context"`
	Position    int                `json:"position"`
	MatchedText string             `json:"matched_text"`
	Timestamp   time.Time          `json:"timestamp"`
}

// InfrastructurePattern represents infrastructure-related work patterns
type InfrastructurePattern struct {
	Type        string    `json:"type"`        // "terraform", "aws", "kubernetes", etc.
	Action      string    `json:"action"`      // "deploy", "configure", "update", etc.
	Component   string    `json:"component"`   // specific resource or service
	Status      string    `json:"status"`      // "completed", "in-progress", "blocked"
	Confidence  float64   `json:"confidence"`
	Context     string    `json:"context"`
	Timestamp   time.Time `json:"timestamp"`
}

// DeploymentPattern represents deployment-related activities
type DeploymentPattern struct {
	Type        string    `json:"type"`        // "deploy", "rollback", "prepare", "validate"
	Environment string    `json:"environment"` // "production", "staging", "development"
	Status      string    `json:"status"`
	Component   string    `json:"component"`
	Confidence  float64   `json:"confidence"`
	Context     string    `json:"context"`
	Timestamp   time.Time `json:"timestamp"`
}

// DevelopmentPattern represents development workflow activities
type DevelopmentPattern struct {
	Type       string    `json:"type"`       // "code_review", "testing", "bug_fix", "feature"
	Action     string    `json:"action"`     // "create", "review", "merge", "fix"
	Status     string    `json:"status"`
	Confidence float64   `json:"confidence"`
	Context    string    `json:"context"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewTechnicalPatternMatcher creates a new pattern matcher with comprehensive DevOps patterns
func NewTechnicalPatternMatcher(debug bool) *TechnicalPatternMatcher {
	matcher := &TechnicalPatternMatcher{
		infrastructurePatterns: make(map[string]*PatternDefinition),
		deploymentPatterns:     make(map[string]*PatternDefinition),
		developmentPatterns:    make(map[string]*PatternDefinition),
		databasePatterns:       make(map[string]*PatternDefinition),
		securityPatterns:       make(map[string]*PatternDefinition),
		testingPatterns:        make(map[string]*PatternDefinition),
		debug:                  debug,
	}
	
	matcher.initializePatterns()
	return matcher
}

// initializePatterns sets up the comprehensive DevOps terminology database
func (m *TechnicalPatternMatcher) initializePatterns() {
	// Infrastructure patterns
	m.infrastructurePatterns["terraform"] = &PatternDefinition{
		Name:        "Terraform Infrastructure",
		Keywords:    []string{"terraform", "tf", "spacelift", "infrastructure as code", "iac"},
		Category:    "infrastructure",
		Subcategory: "terraform",
		BaseScore:   0.9,
		Modifiers: map[string]float64{
			"apply":     0.2,
			"plan":      0.15,
			"destroy":   0.25,
			"init":      0.1,
			"validate":  0.1,
			"import":    0.15,
			"state":     0.1,
			"workspace": 0.1,
		},
		Examples: []string{
			"Applied Terraform configuration",
			"Terraform plan shows changes",
			"Updated Terraform modules",
		},
	}
	
	m.infrastructurePatterns["aws"] = &PatternDefinition{
		Name:        "AWS Cloud Services",
		Keywords:    []string{"aws", "amazon web services", "ec2", "s3", "rds", "lambda", "vpc", "ecr", "ecs", "eks"},
		Category:    "infrastructure",
		Subcategory: "aws",
		BaseScore:   0.85,
		Modifiers: map[string]float64{
			"deploy":    0.2,
			"configure": 0.15,
			"setup":     0.15,
			"provision": 0.2,
			"scale":     0.15,
			"monitor":   0.1,
		},
		Examples: []string{
			"Configured AWS VPC endpoints",
			"Deployed to AWS ECS",
			"Set up AWS IAM roles",
		},
	}
	
	m.infrastructurePatterns["kubernetes"] = &PatternDefinition{
		Name:        "Kubernetes Container Orchestration",
		Keywords:    []string{"kubernetes", "k8s", "kubectl", "helm", "pod", "deployment", "service", "ingress", "namespace"},
		Category:    "infrastructure",
		Subcategory: "kubernetes",
		BaseScore:   0.9,
		Modifiers: map[string]float64{
			"deploy":     0.2,
			"scale":      0.15,
			"rollout":    0.2,
			"configure":  0.15,
			"troubleshoot": 0.1,
		},
		Examples: []string{
			"Deployed Kubernetes manifests",
			"Scaled K8s deployment",
			"Updated Helm charts",
		},
	}
	
	// Deployment patterns
	m.deploymentPatterns["deployment"] = &PatternDefinition{
		Name:        "Application Deployment",
		Keywords:    []string{"deploy", "deployment", "release", "rollout", "publish", "ship"},
		Category:    "deployment",
		Subcategory: "application",
		BaseScore:   0.85,
		Modifiers: map[string]float64{
			"production":  0.3,
			"staging":     0.2,
			"development": 0.1,
			"rollback":    0.25,
			"hotfix":      0.3,
			"canary":      0.2,
			"blue-green":  0.2,
		},
		Examples: []string{
			"Deployed to production",
			"Staging deployment completed",
			"Rolled back deployment",
		},
	}
	
	m.deploymentPatterns["cicd"] = &PatternDefinition{
		Name:        "CI/CD Pipeline",
		Keywords:    []string{"ci/cd", "pipeline", "jenkins", "github actions", "gitlab ci", "build", "continuous integration"},
		Category:    "deployment",
		Subcategory: "pipeline",
		BaseScore:   0.8,
		Modifiers: map[string]float64{
			"failed":    0.2,
			"passed":    0.15,
			"triggered": 0.1,
			"fixed":     0.2,
			"optimized": 0.15,
		},
		Examples: []string{
			"CI/CD pipeline triggered",
			"Fixed pipeline failure",
			"Optimized build process",
		},
	}
	
	// Development patterns
	m.developmentPatterns["code_review"] = &PatternDefinition{
		Name:        "Code Review Process",
		Keywords:    []string{"pr", "pull request", "merge request", "code review", "review", "approve", "lgtm"},
		Category:    "development",
		Subcategory: "code_review",
		BaseScore:   0.8,
		Modifiers: map[string]float64{
			"created":   0.15,
			"merged":    0.2,
			"approved":  0.15,
			"reviewed":  0.1,
			"feedback":  0.1,
			"addressed": 0.15,
		},
		Examples: []string{
			"Created pull request",
			"Merged PR after review",
			"Addressed review feedback",
		},
	}
	
	m.developmentPatterns["bug_fix"] = &PatternDefinition{
		Name:        "Bug Fix and Troubleshooting",
		Keywords:    []string{"bug", "fix", "error", "issue", "problem", "troubleshoot", "debug", "resolve"},
		Category:    "development",
		Subcategory: "bug_fix",
		BaseScore:   0.85,
		Modifiers: map[string]float64{
			"critical":   0.3,
			"urgent":     0.25,
			"production": 0.3,
			"hotfix":     0.3,
			"resolved":   0.2,
			"identified": 0.15,
		},
		Examples: []string{
			"Fixed critical bug",
			"Resolved production issue",
			"Applied hotfix",
		},
	}
	
	// Database patterns
	m.databasePatterns["database"] = &PatternDefinition{
		Name:        "Database Operations",
		Keywords:    []string{"database", "db", "sql", "postgresql", "mysql", "mongodb", "migration", "schema"},
		Category:    "database",
		Subcategory: "operations",
		BaseScore:   0.8,
		Modifiers: map[string]float64{
			"migration":   0.2,
			"backup":      0.15,
			"restore":     0.2,
			"optimize":    0.15,
			"permissions": 0.15,
			"index":       0.1,
		},
		Examples: []string{
			"Ran database migration",
			"Updated database permissions",
			"Optimized database queries",
		},
	}
	
	m.databasePatterns["liquibase"] = &PatternDefinition{
		Name:        "Liquibase Database Management",
		Keywords:    []string{"liquibase", "changelog", "changeset", "rollback", "database versioning"},
		Category:    "database",
		Subcategory: "liquibase",
		BaseScore:   0.85,
		Modifiers: map[string]float64{
			"update":   0.2,
			"rollback": 0.25,
			"validate": 0.15,
			"generate": 0.15,
		},
		Examples: []string{
			"Applied Liquibase changes",
			"Generated Liquibase changelog",
			"Validated database schema",
		},
	}
	
	// Security patterns
	m.securityPatterns["authentication"] = &PatternDefinition{
		Name:        "Authentication and Authorization",
		Keywords:    []string{"auth", "authentication", "authorization", "oauth", "oidc", "jwt", "sso", "saml"},
		Category:    "security",
		Subcategory: "authentication",
		BaseScore:   0.85,
		Modifiers: map[string]float64{
			"configure": 0.2,
			"integrate": 0.2,
			"fix":       0.25,
			"setup":     0.15,
			"validate":  0.15,
		},
		Examples: []string{
			"Configured OAuth integration",
			"Set up OIDC authentication",
			"Fixed JWT validation",
		},
	}
	
	m.securityPatterns["secrets"] = &PatternDefinition{
		Name:        "Secrets Management",
		Keywords:    []string{"secrets", "credentials", "api key", "token", "certificate", "ssl", "tls"},
		Category:    "security",
		Subcategory: "secrets",
		BaseScore:   0.9,
		Modifiers: map[string]float64{
			"rotate":    0.2,
			"configure": 0.15,
			"secure":    0.2,
			"encrypt":   0.2,
			"vault":     0.15,
		},
		Examples: []string{
			"Rotated API keys",
			"Configured secrets vault",
			"Updated SSL certificates",
		},
	}
	
	// Testing patterns
	m.testingPatterns["testing"] = &PatternDefinition{
		Name:        "Software Testing",
		Keywords:    []string{"test", "testing", "unit test", "integration test", "e2e", "qa", "validation"},
		Category:    "testing",
		Subcategory: "general",
		BaseScore:   0.75,
		Modifiers: map[string]float64{
			"passed":   0.15,
			"failed":   0.2,
			"created":  0.15,
			"updated":  0.1,
			"automated": 0.15,
		},
		Examples: []string{
			"Tests passed successfully",
			"Created unit tests",
			"Fixed failing tests",
		},
	}
}

// MatchInfrastructurePatterns finds infrastructure patterns in text
func (m *TechnicalPatternMatcher) MatchInfrastructurePatterns(text string) ([]InfrastructurePattern, error) {
	var patterns []InfrastructurePattern
	lowerText := strings.ToLower(text)
	
	for _, patternDef := range m.infrastructurePatterns {
		matches := m.findPatternMatches(lowerText, text, patternDef)
		
		for _, match := range matches {
			pattern := InfrastructurePattern{
				Type:       patternDef.Subcategory,
				Action:     m.extractAction(match.MatchedText),
				Component:  m.extractComponent(match.MatchedText, patternDef.Subcategory),
				Status:     m.extractStatus(match.MatchedText),
				Confidence: match.Confidence,
				Context:    match.Context,
				Timestamp:  time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns, nil
}

// MatchDeploymentPatterns finds deployment patterns in text
func (m *TechnicalPatternMatcher) MatchDeploymentPatterns(text string) ([]DeploymentPattern, error) {
	var patterns []DeploymentPattern
	lowerText := strings.ToLower(text)
	
	for _, patternDef := range m.deploymentPatterns {
		matches := m.findPatternMatches(lowerText, text, patternDef)
		
		for _, match := range matches {
			pattern := DeploymentPattern{
				Type:        m.extractDeploymentType(match.MatchedText),
				Environment: m.extractEnvironment(match.MatchedText),
				Status:      m.extractStatus(match.MatchedText),
				Component:   m.extractComponent(match.MatchedText, "deployment"),
				Confidence:  match.Confidence,
				Context:     match.Context,
				Timestamp:   time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns, nil
}

// MatchDevelopmentPatterns finds development patterns in text
func (m *TechnicalPatternMatcher) MatchDevelopmentPatterns(text string) ([]DevelopmentPattern, error) {
	var patterns []DevelopmentPattern
	lowerText := strings.ToLower(text)
	
	for _, patternDef := range m.developmentPatterns {
		matches := m.findPatternMatches(lowerText, text, patternDef)
		
		for _, match := range matches {
			pattern := DevelopmentPattern{
				Type:       patternDef.Subcategory,
				Action:     m.extractAction(match.MatchedText),
				Status:     m.extractStatus(match.MatchedText),
				Confidence: match.Confidence,
				Context:    match.Context,
				Timestamp:  time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns, nil
}

// findPatternMatches finds all matches for a pattern definition in text
func (m *TechnicalPatternMatcher) findPatternMatches(lowerText, originalText string, patternDef *PatternDefinition) []PatternMatch {
	var matches []PatternMatch
	
	// Check for keyword matches
	for _, keyword := range patternDef.Keywords {
		if strings.Contains(lowerText, keyword) {
			confidence := m.calculateConfidence(lowerText, patternDef, keyword)
			context := m.extractContext(originalText, keyword)
			
			match := PatternMatch{
				Pattern:     patternDef,
				Text:        originalText,
				Confidence:  confidence,
				Context:     context,
				Position:    strings.Index(lowerText, keyword),
				MatchedText: keyword,
				Timestamp:   time.Now(),
			}
			matches = append(matches, match)
		}
	}
	
	return matches
}

// calculateConfidence calculates confidence score for a pattern match
func (m *TechnicalPatternMatcher) calculateConfidence(text string, patternDef *PatternDefinition, matchedKeyword string) float64 {
	confidence := patternDef.BaseScore
	
	// Apply modifiers based on context
	for modifier, boost := range patternDef.Modifiers {
		if strings.Contains(text, modifier) {
			confidence += boost
			if m.debug {
				// Debug logging would go here
			}
		}
	}
	
	// Boost confidence for multiple keyword matches
	keywordCount := 0
	for _, keyword := range patternDef.Keywords {
		if strings.Contains(text, keyword) {
			keywordCount++
		}
	}
	
	if keywordCount > 1 {
		confidence += float64(keywordCount-1) * 0.1
	}
	
	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// extractContext extracts surrounding context for a matched keyword
func (m *TechnicalPatternMatcher) extractContext(text, keyword string) string {
	lowerText := strings.ToLower(text)
	keywordPos := strings.Index(lowerText, keyword)
	
	if keywordPos == -1 {
		return text
	}
	
	// Extract context around the keyword (50 characters before and after)
	start := keywordPos - 50
	if start < 0 {
		start = 0
	}
	
	end := keywordPos + len(keyword) + 50
	if end > len(text) {
		end = len(text)
	}
	
	context := text[start:end]
	
	// Clean up context to word boundaries
	if start > 0 {
		if spaceIndex := strings.Index(context, " "); spaceIndex != -1 {
			context = context[spaceIndex+1:]
		}
	}
	
	if end < len(text) {
		if spaceIndex := strings.LastIndex(context, " "); spaceIndex != -1 {
			context = context[:spaceIndex]
		}
	}
	
	return strings.TrimSpace(context)
}

// extractAction extracts action verbs from matched text
func (m *TechnicalPatternMatcher) extractAction(text string) string {
	lowerText := strings.ToLower(text)
	
	actions := []string{
		"deploy", "configure", "setup", "install", "update", "upgrade",
		"create", "build", "implement", "develop", "fix", "resolve",
		"test", "validate", "verify", "review", "approve", "merge",
		"troubleshoot", "debug", "investigate", "analyze", "optimize",
	}
	
	for _, action := range actions {
		if strings.Contains(lowerText, action) {
			return action
		}
	}
	
	return "unknown"
}

// extractComponent extracts component information from matched text
func (m *TechnicalPatternMatcher) extractComponent(text, category string) string {
	lowerText := strings.ToLower(text)
	
	switch category {
	case "terraform":
		components := []string{"module", "resource", "provider", "state", "workspace"}
		for _, comp := range components {
			if strings.Contains(lowerText, comp) {
				return comp
			}
		}
	case "aws":
		components := []string{"vpc", "ec2", "s3", "rds", "lambda", "ecr", "ecs", "eks", "iam"}
		for _, comp := range components {
			if strings.Contains(lowerText, comp) {
				return comp
			}
		}
	case "kubernetes":
		components := []string{"pod", "deployment", "service", "ingress", "configmap", "secret", "namespace"}
		for _, comp := range components {
			if strings.Contains(lowerText, comp) {
				return comp
			}
		}
	}
	
	return "general"
}

// extractStatus extracts status information from matched text
func (m *TechnicalPatternMatcher) extractStatus(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "done") || strings.Contains(lowerText, "finished") {
		return "completed"
	}
	if strings.Contains(lowerText, "in progress") || strings.Contains(lowerText, "working") || strings.Contains(lowerText, "ongoing") {
		return "in_progress"
	}
	if strings.Contains(lowerText, "blocked") || strings.Contains(lowerText, "stuck") || strings.Contains(lowerText, "waiting") {
		return "blocked"
	}
	if strings.Contains(lowerText, "failed") || strings.Contains(lowerText, "error") || strings.Contains(lowerText, "issue") {
		return "failed"
	}
	if strings.Contains(lowerText, "planned") || strings.Contains(lowerText, "scheduled") || strings.Contains(lowerText, "upcoming") {
		return "planned"
	}
	
	return "unknown"
}

// extractDeploymentType extracts deployment type from matched text
func (m *TechnicalPatternMatcher) extractDeploymentType(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "rollback") {
		return "rollback"
	}
	if strings.Contains(lowerText, "hotfix") {
		return "hotfix"
	}
	if strings.Contains(lowerText, "canary") {
		return "canary"
	}
	if strings.Contains(lowerText, "blue-green") || strings.Contains(lowerText, "blue green") {
		return "blue_green"
	}
	if strings.Contains(lowerText, "deploy") {
		return "deploy"
	}
	
	return "deploy"
}

// extractEnvironment extracts environment information from matched text
func (m *TechnicalPatternMatcher) extractEnvironment(text string) string {
	lowerText := strings.ToLower(text)
	
	environments := []string{"production", "prod", "staging", "stage", "development", "dev", "test", "testing"}
	for _, env := range environments {
		if strings.Contains(lowerText, env) {
			// Normalize environment names
			switch env {
			case "prod":
				return "production"
			case "stage":
				return "staging"
			case "dev":
				return "development"
			case "testing":
				return "test"
			default:
				return env
			}
		}
	}
	
	return "unknown"
}

// GetPatternStatistics returns statistics about pattern matching
func (m *TechnicalPatternMatcher) GetPatternStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["infrastructure_patterns"] = len(m.infrastructurePatterns)
	stats["deployment_patterns"] = len(m.deploymentPatterns)
	stats["development_patterns"] = len(m.developmentPatterns)
	stats["database_patterns"] = len(m.databasePatterns)
	stats["security_patterns"] = len(m.securityPatterns)
	stats["testing_patterns"] = len(m.testingPatterns)
	
	totalPatterns := len(m.infrastructurePatterns) + len(m.deploymentPatterns) + 
		len(m.developmentPatterns) + len(m.databasePatterns) + 
		len(m.securityPatterns) + len(m.testingPatterns)
	
	stats["total_patterns"] = totalPatterns
	
	return stats
}

// MatchAllPatterns matches all pattern types against text and returns comprehensive results
func (m *TechnicalPatternMatcher) MatchAllPatterns(text string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	
	// Match infrastructure patterns
	infraPatterns, err := m.MatchInfrastructurePatterns(text)
	if err != nil {
		return nil, err
	}
	results["infrastructure"] = infraPatterns
	
	// Match deployment patterns
	deployPatterns, err := m.MatchDeploymentPatterns(text)
	if err != nil {
		return nil, err
	}
	results["deployment"] = deployPatterns
	
	// Match development patterns
	devPatterns, err := m.MatchDevelopmentPatterns(text)
	if err != nil {
		return nil, err
	}
	results["development"] = devPatterns
	
	// Calculate overall confidence
	totalMatches := len(infraPatterns) + len(deployPatterns) + len(devPatterns)
	if totalMatches > 0 {
		var totalConfidence float64
		for _, pattern := range infraPatterns {
			totalConfidence += pattern.Confidence
		}
		for _, pattern := range deployPatterns {
			totalConfidence += pattern.Confidence
		}
		for _, pattern := range devPatterns {
			totalConfidence += pattern.Confidence
		}
		results["overall_confidence"] = totalConfidence / float64(totalMatches)
	} else {
		results["overall_confidence"] = 0.0
	}
	
	results["total_matches"] = totalMatches
	
	return results, nil
}
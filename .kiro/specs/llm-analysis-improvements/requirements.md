# Requirements Document

## Introduction

The my-day CLI tool currently has issues with LLM analysis not correctly summarizing the list of Jira tickets the user was working on. The embedded LLM and overall summarization logic needs improvements to provide more accurate and useful summaries for daily standup reports.

## Requirements

### Requirement 1

**User Story:** As a DevOps team member, I want the LLM to accurately summarize my Jira ticket activity, so that I can quickly understand what I worked on for standup meetings.

#### Acceptance Criteria

1. WHEN the LLM processes multiple Jira issues THEN it SHALL generate individual summaries that capture the key work done on each ticket
2. WHEN the LLM processes comment data THEN it SHALL extract meaningful technical activities and progress updates
3. WHEN generating an overall standup summary THEN it SHALL combine individual ticket summaries into a coherent overview
4. WHEN processing empty or minimal comment data THEN it SHALL provide fallback summaries based on ticket metadata

### Requirement 2

**User Story:** As a user, I want the embedded LLM to understand technical DevOps terminology, so that summaries are relevant and accurate for my work context.

#### Acceptance Criteria

1. WHEN the LLM encounters infrastructure terms THEN it SHALL recognize and properly categorize them (Terraform, AWS, Kubernetes, etc.)
2. WHEN processing deployment-related activities THEN it SHALL identify and highlight deployment status and readiness
3. WHEN analyzing code review activities THEN it SHALL summarize PR creation, merging, and review activities
4. WHEN encountering database work THEN it SHALL identify configuration, migration, and permission-related tasks

### Requirement 3

**User Story:** As a user, I want the LLM to prioritize recent and important activities, so that my standup summary focuses on the most relevant work.

#### Acceptance Criteria

1. WHEN multiple activities are present THEN it SHALL prioritize completed work over planned work
2. WHEN processing time-sensitive activities THEN it SHALL highlight deployment-ready items and blockers
3. WHEN generating summaries THEN it SHALL limit output to the most important 3-5 key activities
4. WHEN encountering duplicate or similar activities THEN it SHALL consolidate them into single summary points

### Requirement 4

**User Story:** As a user, I want debugging capabilities for LLM analysis, so that I can understand why summaries are incorrect or missing.

#### Acceptance Criteria

1. WHEN LLM analysis fails THEN it SHALL provide detailed error information
2. WHEN running in verbose mode THEN it SHALL show intermediate processing steps
3. WHEN summaries seem incorrect THEN it SHALL provide a way to inspect the input data being processed
4. WHEN using different LLM modes THEN it SHALL clearly indicate which mode is active and working

### Requirement 5

**User Story:** As a user, I want the LLM to handle edge cases gracefully, so that the tool remains reliable even with unusual data.

#### Acceptance Criteria

1. WHEN comment text is empty or malformed THEN it SHALL provide meaningful fallback summaries
2. WHEN API calls fail THEN it SHALL degrade gracefully to basic summaries
3. WHEN processing very long comments THEN it SHALL extract key points without truncating important information
4. WHEN encountering non-English text THEN it SHALL handle it appropriately without errors
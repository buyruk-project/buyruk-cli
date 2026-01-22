package models

import (
	"fmt"
	"strconv"
	"strings"
)

// Issue represents a task or bug issue
type Issue struct {
	ID          string   `json:"id"`                    // Required: e.g., "CORE-12"
	Type        string   `json:"type"`                  // Required: "task" or "bug"
	Title       string   `json:"title"`                 // Required
	Status      string   `json:"status"`                // Required: TODO, DOING, DONE
	Priority    string   `json:"priority,omitempty"`    // Optional: LOW, MEDIUM, HIGH, CRITICAL
	Description string   `json:"description,omitempty"` // Optional: Markdown
	PRs         []string `json:"prs,omitempty"`         // Optional: Array of PR URLs
	BlockedBy   []string `json:"blocked_by,omitempty"`  // Optional: Array of issue IDs
	EpicID      string   `json:"epic_id,omitempty"`     // Optional: Link to epic
	CreatedAt   string   `json:"created_at,omitempty"`  // ISO 8601 timestamp
	UpdatedAt   string   `json:"updated_at,omitempty"`  // ISO 8601 timestamp
}

// Validate validates the Issue struct
// ID, Type, and Status are optional (can be auto-generated/defaulted during creation)
// Only Title is required
func (i *Issue) Validate() error {
	// Title is the only required field
	if i.Title == "" {
		return fmt.Errorf("models: issue title is required")
	}

	// Validate Type if provided
	if i.Type != "" && !IsValidType(i.Type) {
		return fmt.Errorf("models: invalid type %q", i.Type)
	}

	// Validate Status if provided
	if i.Status != "" && !IsValidStatus(i.Status) {
		return fmt.Errorf("models: invalid status %q", i.Status)
	}

	// Validate Priority if provided
	if i.Priority != "" && !IsValidPriority(i.Priority) {
		return fmt.Errorf("models: invalid priority %q", i.Priority)
	}

	return nil
}

// AddDependency adds a dependency (blocked by) to the issue
func (i *Issue) AddDependency(issueID string) {
	if !contains(i.BlockedBy, issueID) {
		i.BlockedBy = append(i.BlockedBy, issueID)
	}
}

// RemoveDependency removes a dependency from the issue
func (i *Issue) RemoveDependency(issueID string) {
	i.BlockedBy = remove(i.BlockedBy, issueID)
}

// AddPR adds a PR URL to the issue
func (i *Issue) AddPR(url string) {
	if !contains(i.PRs, url) {
		i.PRs = append(i.PRs, url)
	}
}

// RemovePR removes a PR URL from the issue
func (i *Issue) RemovePR(url string) {
	i.PRs = remove(i.PRs, url)
}

// Epic represents an epic that groups multiple issues
type Epic struct {
	ID          string `json:"id"`                    // Required: e.g., "E-1"
	Title       string `json:"title"`                 // Required
	Description string `json:"description,omitempty"` // Optional: Markdown
	Status      string `json:"status,omitempty"`      // Optional: TODO, DOING, DONE
	CreatedAt   string `json:"created_at,omitempty"`  // ISO 8601 timestamp
	UpdatedAt   string `json:"updated_at,omitempty"`  // ISO 8601 timestamp
}

// Validate validates the Epic struct
func (e *Epic) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("models: epic ID is required")
	}
	if e.Title == "" {
		return fmt.Errorf("models: epic title is required")
	}
	if e.Status != "" && !IsValidStatus(e.Status) {
		return fmt.Errorf("models: invalid status %q", e.Status)
	}
	return nil
}

// IndexEntry represents a single entry in the project index
type IndexEntry struct {
	ID     string `json:"id"`                // Issue ID: e.g., "CORE-12"
	Title  string `json:"title"`             // Issue title
	Status string `json:"status"`            // Issue status
	Type   string `json:"type"`              // Issue type
	EpicID string `json:"epic_id,omitempty"` // Optional epic link
}

// ProjectIndex represents the index of all issues in a project
type ProjectIndex struct {
	ProjectKey  string       `json:"project_key"`            // Required: e.g., "CORE"
	ProjectName string       `json:"project_name,omitempty"` // Optional
	Issues      []IndexEntry `json:"issues"`                 // Array of index entries
	CreatedAt   string       `json:"created_at,omitempty"`   // ISO 8601
	UpdatedAt   string       `json:"updated_at,omitempty"`   // ISO 8601
}

// AddIssue adds an issue to the project index
func (idx *ProjectIndex) AddIssue(issue *Issue) {
	entry := IndexEntry{
		ID:     issue.ID,
		Title:  issue.Title,
		Status: issue.Status,
		Type:   issue.Type,
		EpicID: issue.EpicID,
	}

	// Remove existing entry if present
	idx.RemoveIssue(issue.ID)

	// Add new entry
	idx.Issues = append(idx.Issues, entry)
}

// RemoveIssue removes an issue from the project index
func (idx *ProjectIndex) RemoveIssue(issueID string) {
	idx.Issues = removeIndexEntry(idx.Issues, issueID)
}

// FindIssue finds an issue in the project index by ID
func (idx *ProjectIndex) FindIssue(issueID string) *IndexEntry {
	for i := range idx.Issues {
		if idx.Issues[i].ID == issueID {
			return &idx.Issues[i]
		}
	}
	return nil
}

// Validate validates the ProjectIndex struct
func (idx *ProjectIndex) Validate() error {
	if idx.ProjectKey == "" {
		return fmt.Errorf("models: project key is required")
	}

	// Validate all index entries
	for i, entry := range idx.Issues {
		if entry.ID == "" {
			return fmt.Errorf("models: index entry %d has empty ID", i)
		}
		if !IsValidStatus(entry.Status) {
			return fmt.Errorf("models: index entry %s has invalid status %q", entry.ID, entry.Status)
		}
		if entry.Type != "" && !IsValidType(entry.Type) {
			return fmt.Errorf("models: index entry %s has invalid type %q", entry.ID, entry.Type)
		}
	}

	return nil
}

// Project represents a project
type Project struct {
	Key         string `json:"key"`                   // Required: e.g., "CORE"
	Name        string `json:"name"`                  // Optional
	Description string `json:"description,omitempty"` // Optional
	CreatedAt   string `json:"created_at,omitempty"`  // ISO 8601
	UpdatedAt   string `json:"updated_at,omitempty"`  // ISO 8601
}

// Validate validates the Project struct
func (p *Project) Validate() error {
	if p.Key == "" {
		return fmt.Errorf("models: project key is required")
	}

	// Validate key format (uppercase alphanumeric, no spaces, allows hyphens)
	if !isValidProjectKey(p.Key) {
		return fmt.Errorf("models: invalid project key format %q", p.Key)
	}

	return nil
}

// isValidProjectKey validates that the project key is uppercase alphanumeric or hyphen
func isValidProjectKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, r := range key {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// Helper functions

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// remove removes a string from a string slice
func remove(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// removeIndexEntry removes an IndexEntry from a slice by ID
func removeIndexEntry(entries []IndexEntry, id string) []IndexEntry {
	result := []IndexEntry{}
	for _, e := range entries {
		if e.ID != id {
			result = append(result, e)
		}
	}
	return result
}

// GenerateIssueID generates an issue ID from project key and sequence number
func GenerateIssueID(projectKey string, sequence int) string {
	return fmt.Sprintf("%s-%d", projectKey, sequence)
}

// ParseIssueID parses an issue ID into project key and sequence number
// Supports project keys with hyphens by splitting from the right (last hyphen)
func ParseIssueID(id string) (projectKey string, sequence int, err error) {
	// Find the last hyphen to support project keys with hyphens
	lastHyphen := strings.LastIndex(id, "-")
	if lastHyphen == -1 {
		return "", 0, fmt.Errorf("models: invalid issue ID format %q", id)
	}

	projectKey = id[:lastHyphen]
	sequenceStr := id[lastHyphen+1:]

	// Validate project key is not empty
	if projectKey == "" {
		return "", 0, fmt.Errorf("models: invalid issue ID format %q", id)
	}

	// Validate sequence string is not empty
	if sequenceStr == "" {
		return "", 0, fmt.Errorf("models: invalid issue ID format %q", id)
	}

	// Validate sequence contains only digits (no hyphens or other characters)
	for _, r := range sequenceStr {
		if r < '0' || r > '9' {
			return "", 0, fmt.Errorf("models: invalid sequence in ID %q: sequence must be numeric", id)
		}
	}

	sequence, err = strconv.Atoi(sequenceStr)
	if err != nil {
		return "", 0, fmt.Errorf("models: invalid sequence in ID %q: %w", id, err)
	}

	return projectKey, sequence, nil
}

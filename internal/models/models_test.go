package models

import (
	"encoding/json"
	"testing"
	"time"
)

// Test Constants

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"TODO", StatusTODO, true},
		{"DOING", StatusDOING, true},
		{"DONE", StatusDONE, true},
		{"invalid", "INVALID", false},
		{"empty", "", false},
		{"lowercase", "todo", false},
		{"mixed case", "Doing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidStatus(tt.status)
			if got != tt.expected {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		expected bool
	}{
		{"LOW", PriorityLOW, true},
		{"MEDIUM", PriorityMEDIUM, true},
		{"HIGH", PriorityHIGH, true},
		{"CRITICAL", PriorityCRITICAL, true},
		{"invalid", "INVALID", false},
		{"empty", "", false},
		{"lowercase", "low", false},
		{"mixed case", "Medium", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPriority(tt.priority)
			if got != tt.expected {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tt.priority, got, tt.expected)
			}
		})
	}
}

func TestIsValidType(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		expected bool
	}{
		{"task", TypeTask, true},
		{"bug", TypeBug, true},
		{"epic", TypeEpic, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"uppercase", "TASK", false},
		{"mixed case", "Task", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidType(tt.typ)
			if got != tt.expected {
				t.Errorf("IsValidType(%q) = %v, want %v", tt.typ, got, tt.expected)
			}
		})
	}
}

// Test Issue Model

func TestIssue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		issue   *Issue
		wantErr bool
	}{
		{
			name: "valid issue",
			issue: &Issue{
				ID:     "CORE-12",
				Type:   TypeTask,
				Title:  "Test Issue",
				Status: StatusTODO,
			},
			wantErr: false,
		},
		{
			name: "valid issue with priority",
			issue: &Issue{
				ID:       "CORE-12",
				Type:     TypeTask,
				Title:    "Test Issue",
				Status:   StatusTODO,
				Priority: PriorityHIGH,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			issue: &Issue{
				Type:   TypeTask,
				Title:  "Test Issue",
				Status: StatusTODO,
			},
			wantErr: true,
		},
		{
			name: "missing title",
			issue: &Issue{
				ID:     "CORE-12",
				Type:   TypeTask,
				Status: StatusTODO,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			issue: &Issue{
				ID:     "CORE-12",
				Type:   TypeTask,
				Title:  "Test Issue",
				Status: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			issue: &Issue{
				ID:       "CORE-12",
				Type:     TypeTask,
				Title:    "Test Issue",
				Status:   StatusTODO,
				Priority: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			issue: &Issue{
				ID:     "CORE-12",
				Type:   "invalid",
				Title:  "Test Issue",
				Status: StatusTODO,
			},
			wantErr: true,
		},
		{
			name: "empty type is invalid",
			issue: &Issue{
				ID:     "CORE-12",
				Type:   "",
				Title:  "Test Issue",
				Status: StatusTODO,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.issue.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Issue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIssue_AddDependency(t *testing.T) {
	issue := &Issue{
		ID:     "CORE-12",
		Type:   TypeTask,
		Title:  "Test Issue",
		Status: StatusTODO,
	}

	// Add first dependency
	issue.AddDependency("CORE-10")
	if len(issue.BlockedBy) != 1 {
		t.Errorf("AddDependency() should add one dependency, got %d", len(issue.BlockedBy))
	}
	if issue.BlockedBy[0] != "CORE-10" {
		t.Errorf("AddDependency() added wrong ID, got %q, want CORE-10", issue.BlockedBy[0])
	}

	// Add second dependency
	issue.AddDependency("CORE-11")
	if len(issue.BlockedBy) != 2 {
		t.Errorf("AddDependency() should add second dependency, got %d", len(issue.BlockedBy))
	}

	// Try to add duplicate - should not add
	issue.AddDependency("CORE-10")
	if len(issue.BlockedBy) != 2 {
		t.Errorf("AddDependency() should not add duplicate, got %d dependencies", len(issue.BlockedBy))
	}
}

func TestIssue_RemoveDependency(t *testing.T) {
	issue := &Issue{
		ID:        "CORE-12",
		Type:      TypeTask,
		Title:     "Test Issue",
		Status:    StatusTODO,
		BlockedBy: []string{"CORE-10", "CORE-11", "CORE-12"},
	}

	// Remove middle dependency
	issue.RemoveDependency("CORE-11")
	if len(issue.BlockedBy) != 2 {
		t.Errorf("RemoveDependency() should remove one dependency, got %d", len(issue.BlockedBy))
	}
	if contains(issue.BlockedBy, "CORE-11") {
		t.Error("RemoveDependency() should remove CORE-11")
	}

	// Remove first dependency
	issue.RemoveDependency("CORE-10")
	if len(issue.BlockedBy) != 1 {
		t.Errorf("RemoveDependency() should remove another dependency, got %d", len(issue.BlockedBy))
	}

	// Remove non-existent dependency - should not error
	issue.RemoveDependency("CORE-99")
	if len(issue.BlockedBy) != 1 {
		t.Errorf("RemoveDependency() should not affect count for non-existent, got %d", len(issue.BlockedBy))
	}
}

func TestIssue_AddPR(t *testing.T) {
	issue := &Issue{
		ID:     "CORE-12",
		Type:   TypeTask,
		Title:  "Test Issue",
		Status: StatusTODO,
	}

	// Add first PR
	issue.AddPR("https://github.com/example/repo/pull/1")
	if len(issue.PRs) != 1 {
		t.Errorf("AddPR() should add one PR, got %d", len(issue.PRs))
	}

	// Add second PR
	issue.AddPR("https://github.com/example/repo/pull/2")
	if len(issue.PRs) != 2 {
		t.Errorf("AddPR() should add second PR, got %d", len(issue.PRs))
	}

	// Try to add duplicate - should not add
	issue.AddPR("https://github.com/example/repo/pull/1")
	if len(issue.PRs) != 2 {
		t.Errorf("AddPR() should not add duplicate, got %d PRs", len(issue.PRs))
	}
}

// Test Epic Model

func TestEpic_Validate(t *testing.T) {
	tests := []struct {
		name    string
		epic    *Epic
		wantErr bool
	}{
		{
			name: "valid epic",
			epic: &Epic{
				ID:    "E-1",
				Title: "Test Epic",
			},
			wantErr: false,
		},
		{
			name: "valid epic with status",
			epic: &Epic{
				ID:     "E-1",
				Title:  "Test Epic",
				Status: StatusDOING,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			epic: &Epic{
				Title: "Test Epic",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			epic: &Epic{
				ID: "E-1",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			epic: &Epic{
				ID:     "E-1",
				Title:  "Test Epic",
				Status: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "empty status is valid",
			epic: &Epic{
				ID:     "E-1",
				Title:  "Test Epic",
				Status: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.epic.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Epic.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test ProjectIndex Model

func TestProjectIndex_AddIssue(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey: "CORE",
		Issues:     []IndexEntry{},
	}

	issue := &Issue{
		ID:     "CORE-12",
		Type:   TypeTask,
		Title:  "Test Issue",
		Status: StatusTODO,
		EpicID: "E-1",
	}

	// Add issue
	idx.AddIssue(issue)
	if len(idx.Issues) != 1 {
		t.Errorf("AddIssue() should add one issue, got %d", len(idx.Issues))
	}

	entry := idx.Issues[0]
	if entry.ID != "CORE-12" {
		t.Errorf("AddIssue() entry ID = %q, want CORE-12", entry.ID)
	}
	if entry.Title != "Test Issue" {
		t.Errorf("AddIssue() entry Title = %q, want Test Issue", entry.Title)
	}
	if entry.Status != StatusTODO {
		t.Errorf("AddIssue() entry Status = %q, want %s", entry.Status, StatusTODO)
	}
	if entry.Type != TypeTask {
		t.Errorf("AddIssue() entry Type = %q, want %s", entry.Type, TypeTask)
	}
	if entry.EpicID != "E-1" {
		t.Errorf("AddIssue() entry EpicID = %q, want E-1", entry.EpicID)
	}

	// Update issue - should replace, not duplicate
	issue.Title = "Updated Issue"
	issue.Status = StatusDOING
	idx.AddIssue(issue)
	if len(idx.Issues) != 1 {
		t.Errorf("AddIssue() should replace existing issue, got %d issues", len(idx.Issues))
	}
	if idx.Issues[0].Title != "Updated Issue" {
		t.Errorf("AddIssue() should update title, got %q", idx.Issues[0].Title)
	}
	if idx.Issues[0].Status != StatusDOING {
		t.Errorf("AddIssue() should update status, got %q", idx.Issues[0].Status)
	}
}

func TestProjectIndex_RemoveIssue(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey: "CORE",
		Issues: []IndexEntry{
			{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
			{ID: "CORE-11", Title: "Issue 11", Status: StatusDOING, Type: TypeBug},
			{ID: "CORE-12", Title: "Issue 12", Status: StatusDONE, Type: TypeTask},
		},
	}

	// Remove middle issue
	idx.RemoveIssue("CORE-11")
	if len(idx.Issues) != 2 {
		t.Errorf("RemoveIssue() should remove one issue, got %d", len(idx.Issues))
	}
	if idx.FindIssue("CORE-11") != nil {
		t.Error("RemoveIssue() should remove CORE-11")
	}

	// Remove first issue
	idx.RemoveIssue("CORE-10")
	if len(idx.Issues) != 1 {
		t.Errorf("RemoveIssue() should remove another issue, got %d", len(idx.Issues))
	}

	// Remove non-existent issue - should not error
	idx.RemoveIssue("CORE-99")
	if len(idx.Issues) != 1 {
		t.Errorf("RemoveIssue() should not affect count for non-existent, got %d", len(idx.Issues))
	}
}

func TestProjectIndex_FindIssue(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey: "CORE",
		Issues: []IndexEntry{
			{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
			{ID: "CORE-11", Title: "Issue 11", Status: StatusDOING, Type: TypeBug},
		},
	}

	// Find existing issue
	entry := idx.FindIssue("CORE-10")
	if entry == nil {
		t.Fatal("FindIssue() should find CORE-10")
	}
	if entry.Title != "Issue 10" {
		t.Errorf("FindIssue() entry Title = %q, want Issue 10", entry.Title)
	}

	// Find another existing issue
	entry = idx.FindIssue("CORE-11")
	if entry == nil {
		t.Fatal("FindIssue() should find CORE-11")
	}

	// Find non-existent issue
	entry = idx.FindIssue("CORE-99")
	if entry != nil {
		t.Error("FindIssue() should return nil for non-existent issue")
	}
}

func TestProjectIndex_Validate(t *testing.T) {
	tests := []struct {
		name    string
		index   *ProjectIndex
		wantErr bool
	}{
		{
			name: "valid index",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues: []IndexEntry{
					{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
				},
			},
			wantErr: false,
		},
		{
			name: "valid index with multiple issues",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues: []IndexEntry{
					{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
					{ID: "CORE-11", Title: "Issue 11", Status: StatusDOING, Type: TypeBug},
				},
			},
			wantErr: false,
		},
		{
			name: "valid index with empty issues",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues:     []IndexEntry{},
			},
			wantErr: false,
		},
		{
			name: "missing project key",
			index: &ProjectIndex{
				Issues: []IndexEntry{
					{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
				},
			},
			wantErr: true,
		},
		{
			name: "index entry with empty ID",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues: []IndexEntry{
					{ID: "", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
				},
			},
			wantErr: true,
		},
		{
			name: "index entry with invalid status",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues: []IndexEntry{
					{ID: "CORE-10", Title: "Issue 10", Status: "INVALID", Type: TypeTask},
				},
			},
			wantErr: true,
		},
		{
			name: "index entry with invalid type",
			index: &ProjectIndex{
				ProjectKey: "CORE",
				Issues: []IndexEntry{
					{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: "INVALID"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.index.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectIndex.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test Project Model

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		wantErr bool
	}{
		{
			name: "valid project",
			project: &Project{
				Key: "CORE",
			},
			wantErr: false,
		},
		{
			name: "valid project with name",
			project: &Project{
				Key:  "CORE",
				Name: "Core Project",
			},
			wantErr: false,
		},
		{
			name: "valid project with hyphen",
			project: &Project{
				Key: "TEST-PROJ",
			},
			wantErr: false,
		},
		{
			name: "valid project with numbers",
			project: &Project{
				Key: "PROJ123",
			},
			wantErr: false,
		},
		{
			name: "missing key",
			project: &Project{
				Name: "Test Project",
			},
			wantErr: true,
		},
		{
			name: "lowercase key",
			project: &Project{
				Key: "core",
			},
			wantErr: true,
		},
		{
			name: "mixed case key",
			project: &Project{
				Key: "Core",
			},
			wantErr: true,
		},
		{
			name: "key with underscore",
			project: &Project{
				Key: "TEST_PROJ",
			},
			wantErr: true,
		},
		{
			name: "key with space",
			project: &Project{
				Key: "TEST PROJ",
			},
			wantErr: true,
		},
		{
			name: "empty key",
			project: &Project{
				Key: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Project.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test Helper Functions

func TestGenerateIssueID(t *testing.T) {
	tests := []struct {
		name       string
		projectKey string
		sequence   int
		expected   string
	}{
		{"simple", "CORE", 12, "CORE-12"},
		{"single digit", "TEST", 1, "TEST-1"},
		{"large number", "PROJ", 9999, "PROJ-9999"},
		{"zero", "TEST", 0, "TEST-0"},
		{"with hyphen", "TEST-PROJ", 5, "TEST-PROJ-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateIssueID(tt.projectKey, tt.sequence)
			if got != tt.expected {
				t.Errorf("GenerateIssueID(%q, %d) = %q, want %q", tt.projectKey, tt.sequence, got, tt.expected)
			}
		})
	}
}

func TestParseIssueID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		wantKey      string
		wantSequence int
		wantErr      bool
	}{
		{"simple", "CORE-12", "CORE", 12, false},
		{"single digit", "TEST-1", "TEST", 1, false},
		{"large number", "PROJ-9999", "PROJ", 9999, false},
		{"with hyphen in key", "TEST-PROJ-5", "TEST-PROJ", 5, false},
		{"multiple separators with hyphen key", "CORE-12-34", "CORE-12", 34, false},
		{"no separator", "CORE12", "", 0, true},
		{"non-numeric sequence", "CORE-abc", "", 0, true},
		{"empty", "", "", 0, true},
		{"only key", "CORE", "", 0, true},
		{"only sequence", "-12", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotSequence, err := ParseIssueID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIssueID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotKey != tt.wantKey {
					t.Errorf("ParseIssueID(%q) key = %q, want %q", tt.id, gotKey, tt.wantKey)
				}
				if gotSequence != tt.wantSequence {
					t.Errorf("ParseIssueID(%q) sequence = %d, want %d", tt.id, gotSequence, tt.wantSequence)
				}
			}
		})
	}
}

// Test JSON Serialization

func TestIssue_JSON(t *testing.T) {
	issue := &Issue{
		ID:          "CORE-12",
		Type:        TypeTask,
		Title:       "Test Issue",
		Status:      StatusTODO,
		Priority:    PriorityHIGH,
		Description: "Test description",
		PRs:         []string{"https://github.com/example/repo/pull/1"},
		BlockedBy:   []string{"CORE-10"},
		EpicID:      "E-1",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Marshal
	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled Issue
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != issue.ID {
		t.Errorf("Unmarshaled ID = %q, want %q", unmarshaled.ID, issue.ID)
	}
	if unmarshaled.Type != issue.Type {
		t.Errorf("Unmarshaled Type = %q, want %q", unmarshaled.Type, issue.Type)
	}
	if unmarshaled.Title != issue.Title {
		t.Errorf("Unmarshaled Title = %q, want %q", unmarshaled.Title, issue.Title)
	}
	if unmarshaled.Status != issue.Status {
		t.Errorf("Unmarshaled Status = %q, want %q", unmarshaled.Status, issue.Status)
	}
	if unmarshaled.Priority != issue.Priority {
		t.Errorf("Unmarshaled Priority = %q, want %q", unmarshaled.Priority, issue.Priority)
	}
	if len(unmarshaled.PRs) != 1 || unmarshaled.PRs[0] != issue.PRs[0] {
		t.Errorf("Unmarshaled PRs = %v, want %v", unmarshaled.PRs, issue.PRs)
	}
	if len(unmarshaled.BlockedBy) != 1 || unmarshaled.BlockedBy[0] != issue.BlockedBy[0] {
		t.Errorf("Unmarshaled BlockedBy = %v, want %v", unmarshaled.BlockedBy, issue.BlockedBy)
	}
}

func TestIssue_JSON_EmptyFields(t *testing.T) {
	issue := &Issue{
		ID:     "CORE-12",
		Type:   TypeTask,
		Title:  "Test Issue",
		Status: StatusTODO,
		// Optional fields are empty
	}

	// Marshal
	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Verify omitempty works - should not include empty optional fields
	dataStr := string(data)
	if contains([]string{dataStr}, `"priority"`) {
		t.Error("JSON should not include empty priority field")
	}
	if contains([]string{dataStr}, `"description"`) {
		t.Error("JSON should not include empty description field")
	}
	if contains([]string{dataStr}, `"prs"`) {
		t.Error("JSON should not include empty prs field")
	}
	if contains([]string{dataStr}, `"blocked_by"`) {
		t.Error("JSON should not include empty blocked_by field")
	}
	if contains([]string{dataStr}, `"epic_id"`) {
		t.Error("JSON should not include empty epic_id field")
	}
}

func TestEpic_JSON(t *testing.T) {
	epic := &Epic{
		ID:          "E-1",
		Title:       "Test Epic",
		Description: "Test description",
		Status:      StatusDOING,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Marshal
	data, err := json.MarshalIndent(epic, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled Epic
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != epic.ID {
		t.Errorf("Unmarshaled ID = %q, want %q", unmarshaled.ID, epic.ID)
	}
	if unmarshaled.Title != epic.Title {
		t.Errorf("Unmarshaled Title = %q, want %q", unmarshaled.Title, epic.Title)
	}
	if unmarshaled.Status != epic.Status {
		t.Errorf("Unmarshaled Status = %q, want %q", unmarshaled.Status, epic.Status)
	}
}

func TestProjectIndex_JSON(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey:  "CORE",
		ProjectName: "Core Project",
		Issues: []IndexEntry{
			{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
			{ID: "CORE-11", Title: "Issue 11", Status: StatusDOING, Type: TypeBug, EpicID: "E-1"},
		},
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Marshal
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled ProjectIndex
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.ProjectKey != idx.ProjectKey {
		t.Errorf("Unmarshaled ProjectKey = %q, want %q", unmarshaled.ProjectKey, idx.ProjectKey)
	}
	if len(unmarshaled.Issues) != 2 {
		t.Errorf("Unmarshaled Issues length = %d, want 2", len(unmarshaled.Issues))
	}
	if unmarshaled.Issues[0].ID != "CORE-10" {
		t.Errorf("Unmarshaled Issues[0].ID = %q, want CORE-10", unmarshaled.Issues[0].ID)
	}
}

func TestProject_JSON(t *testing.T) {
	project := &Project{
		Key:         "CORE",
		Name:        "Core Project",
		Description: "Test description",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Marshal
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled Project
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.Key != project.Key {
		t.Errorf("Unmarshaled Key = %q, want %q", unmarshaled.Key, project.Key)
	}
	if unmarshaled.Name != project.Name {
		t.Errorf("Unmarshaled Name = %q, want %q", unmarshaled.Name, project.Name)
	}
}

// Test Validation Methods (direct calls)

func TestValidateIssue(t *testing.T) {
	issue := &Issue{
		ID:     "CORE-12",
		Type:   TypeTask,
		Title:  "Test Issue",
		Status: StatusTODO,
	}

	err := issue.Validate()
	if err != nil {
		t.Errorf("Issue.Validate() should succeed for valid issue, got: %v", err)
	}

	invalidIssue := &Issue{
		ID: "",
	}
	err = invalidIssue.Validate()
	if err == nil {
		t.Error("Issue.Validate() should fail for invalid issue")
	}
}

func TestValidateEpic(t *testing.T) {
	epic := &Epic{
		ID:    "E-1",
		Title: "Test Epic",
	}

	err := epic.Validate()
	if err != nil {
		t.Errorf("Epic.Validate() should succeed for valid epic, got: %v", err)
	}

	invalidEpic := &Epic{
		ID: "",
	}
	err = invalidEpic.Validate()
	if err == nil {
		t.Error("Epic.Validate() should fail for invalid epic")
	}
}

func TestValidateProject(t *testing.T) {
	project := &Project{
		Key: "CORE",
	}

	err := project.Validate()
	if err != nil {
		t.Errorf("Project.Validate() should succeed for valid project, got: %v", err)
	}

	invalidProject := &Project{
		Key: "",
	}
	err = invalidProject.Validate()
	if err == nil {
		t.Error("Project.Validate() should fail for invalid project")
	}
}

func TestValidateProjectIndex(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey: "CORE",
		Issues: []IndexEntry{
			{ID: "CORE-10", Title: "Issue 10", Status: StatusTODO, Type: TypeTask},
		},
	}

	err := idx.Validate()
	if err != nil {
		t.Errorf("ProjectIndex.Validate() should succeed for valid index, got: %v", err)
	}

	invalidIdx := &ProjectIndex{
		ProjectKey: "",
	}
	err = invalidIdx.Validate()
	if err == nil {
		t.Error("ProjectIndex.Validate() should fail for invalid index")
	}
}

// Test Edge Cases

func TestIssue_EmptySlices(t *testing.T) {
	issue := &Issue{
		ID:        "CORE-12",
		Type:      TypeTask,
		Title:     "Test Issue",
		Status:    StatusTODO,
		PRs:       []string{},
		BlockedBy: []string{},
	}

	// Should validate successfully
	if err := issue.Validate(); err != nil {
		t.Errorf("Issue with empty slices should validate, got: %v", err)
	}

	// Should handle operations on empty slices
	issue.AddDependency("CORE-10")
	if len(issue.BlockedBy) != 1 {
		t.Errorf("AddDependency() should work on empty slice, got %d dependencies", len(issue.BlockedBy))
	}

	issue.AddPR("https://example.com/pr/1")
	if len(issue.PRs) != 1 {
		t.Errorf("AddPR() should work on empty slice, got %d PRs", len(issue.PRs))
	}
}

func TestProjectIndex_EmptyIssues(t *testing.T) {
	idx := &ProjectIndex{
		ProjectKey: "CORE",
		Issues:     []IndexEntry{},
	}

	// Should validate successfully
	if err := idx.Validate(); err != nil {
		t.Errorf("ProjectIndex with empty issues should validate, got: %v", err)
	}

	// Should handle operations on empty index
	entry := idx.FindIssue("CORE-10")
	if entry != nil {
		t.Error("FindIssue() should return nil for empty index")
	}

	idx.RemoveIssue("CORE-10")
	if len(idx.Issues) != 0 {
		t.Errorf("RemoveIssue() should not affect empty index, got %d issues", len(idx.Issues))
	}
}

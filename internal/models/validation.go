package models

// ValidateIssue validates an Issue struct
func ValidateIssue(issue *Issue) error {
	return issue.Validate()
}

// ValidateEpic validates an Epic struct
func ValidateEpic(epic *Epic) error {
	return epic.Validate()
}

// ValidateProject validates a Project struct
func ValidateProject(project *Project) error {
	return project.Validate()
}

// ValidateProjectIndex validates a ProjectIndex struct
func ValidateProjectIndex(index *ProjectIndex) error {
	return index.Validate()
}

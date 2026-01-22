package models

// Status constants
const (
	StatusTODO  = "TODO"
	StatusDOING = "DOING"
	StatusDONE  = "DONE"
)

// ValidStatuses contains all valid status values
var ValidStatuses = []string{StatusTODO, StatusDOING, StatusDONE}

// Priority constants
const (
	PriorityLOW      = "LOW"
	PriorityMEDIUM   = "MEDIUM"
	PriorityHIGH     = "HIGH"
	PriorityCRITICAL = "CRITICAL"
)

// ValidPriorities contains all valid priority values
var ValidPriorities = []string{
	PriorityLOW,
	PriorityMEDIUM,
	PriorityHIGH,
	PriorityCRITICAL,
}

// Type constants
const (
	TypeTask = "task"
	TypeBug  = "bug"
	TypeEpic = "epic"
)

// ValidTypes contains all valid type values
var ValidTypes = []string{TypeTask, TypeBug, TypeEpic}

// IsValidStatus checks if the given string is a valid status
func IsValidStatus(s string) bool {
	for _, valid := range ValidStatuses {
		if s == valid {
			return true
		}
	}
	return false
}

// IsValidPriority checks if the given string is a valid priority
func IsValidPriority(p string) bool {
	for _, valid := range ValidPriorities {
		if p == valid {
			return true
		}
	}
	return false
}

// IsValidType checks if the given string is a valid type
func IsValidType(t string) bool {
	for _, valid := range ValidTypes {
		if t == valid {
			return true
		}
	}
	return false
}

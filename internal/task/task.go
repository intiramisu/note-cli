package task

import (
	"time"
)

type Priority int

const (
	PriorityNone Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

// PriorityCycle is the order of priority cycling in TUI.
var PriorityCycle = []Priority{PriorityHigh, PriorityMedium, PriorityLow}

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "P3"
	case PriorityMedium:
		return "P2"
	case PriorityHigh:
		return "P1"
	default:
		return ""
	}
}

// CyclePriority cycles through priorities in order or reverse.
func CyclePriority(current Priority, reverse bool) Priority {
	for i, p := range PriorityCycle {
		if p == current {
			next := i + 1
			if reverse {
				next = i - 1 + len(PriorityCycle)
			}
			return PriorityCycle[next%len(PriorityCycle)]
		}
	}
	return PriorityCycle[0]
}

// ParsePriority parses a string into a Priority value.
func ParsePriority(s string) Priority {
	switch s {
	case "1", "high":
		return PriorityHigh
	case "2", "medium":
		return PriorityMedium
	case "3", "low":
		return PriorityLow
	default:
		return PriorityNone
	}
}

type Status int

const (
	StatusPending Status = iota
	StatusDone
)

type Task struct {
	ID          int       `yaml:"id"`
	Description string    `yaml:"description"`
	Priority    Priority  `yaml:"priority"`
	Status      Status    `yaml:"status"`
	NoteID      string    `yaml:"note_id,omitempty"`
	DueDate     time.Time `yaml:"due_date,omitempty"`
	Created     time.Time `yaml:"created"`
	Completed   time.Time `yaml:"completed,omitempty"`
}

func NewTask(id int, description string, priority Priority) *Task {
	return &Task{
		ID:          id,
		Description: description,
		Priority:    priority,
		Status:      StatusPending,
		Created:     time.Now(),
	}
}

func (t *Task) Done() {
	t.Status = StatusDone
	t.Completed = time.Now()
}

func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Task) SetNoteID(noteID string) {
	t.NoteID = noteID
}

func (t *Task) HasNote() bool {
	return t.NoteID != ""
}

func (t *Task) SetDueDate(due time.Time) {
	t.DueDate = due
}

func (t *Task) HasDueDate() bool {
	return !t.DueDate.IsZero()
}

func (t *Task) IsOverdue() bool {
	if !t.HasDueDate() || t.IsDone() {
		return false
	}
	return time.Now().After(t.DueDate)
}

func (t *Task) IsDueSoon(days int) bool {
	if !t.HasDueDate() || t.IsDone() {
		return false
	}
	deadline := time.Now().AddDate(0, 0, days)
	return t.DueDate.Before(deadline) && !t.IsOverdue()
}

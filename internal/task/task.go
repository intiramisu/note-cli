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
	Created     time.Time `yaml:"created"`
	Completed   time.Time `yaml:"completed,omitempty"`
	DueDate     time.Time `yaml:"due_date,omitempty"`
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

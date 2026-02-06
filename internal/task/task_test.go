package task

import (
	"testing"
	"time"
)

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		want     string
	}{
		{PriorityHigh, "P1"},
		{PriorityMedium, "P2"},
		{PriorityLow, "P3"},
		{PriorityNone, ""},
	}

	for _, tt := range tests {
		got := tt.priority.String()
		if got != tt.want {
			t.Errorf("Priority(%d).String() = %q, want %q", tt.priority, got, tt.want)
		}
	}
}

func TestNewTask(t *testing.T) {
	id := 1
	description := "テストタスク"
	priority := PriorityHigh

	before := time.Now()
	task := NewTask(id, description, priority)
	after := time.Now()

	if task.ID != id {
		t.Errorf("ID = %d, want %d", task.ID, id)
	}

	if task.Description != description {
		t.Errorf("Description = %q, want %q", task.Description, description)
	}

	if task.Priority != priority {
		t.Errorf("Priority = %d, want %d", task.Priority, priority)
	}

	if task.Status != StatusPending {
		t.Errorf("Status = %d, want %d", task.Status, StatusPending)
	}

	if task.Created.Before(before) || task.Created.After(after) {
		t.Error("Created time out of range")
	}

	if task.NoteID != "" {
		t.Errorf("NoteID = %q, want empty", task.NoteID)
	}
}

func TestTaskDone(t *testing.T) {
	task := NewTask(1, "テスト", PriorityMedium)

	if task.IsDone() {
		t.Error("New task should not be done")
	}

	before := time.Now()
	task.Done()
	after := time.Now()

	if !task.IsDone() {
		t.Error("Task should be done after Done()")
	}

	if task.Status != StatusDone {
		t.Errorf("Status = %d, want %d", task.Status, StatusDone)
	}

	if task.Completed.Before(before) || task.Completed.After(after) {
		t.Error("Completed time out of range")
	}
}

func TestTaskSetNoteID(t *testing.T) {
	task := NewTask(1, "テスト", PriorityLow)

	if task.HasNote() {
		t.Error("New task should not have note")
	}

	noteID := "会議メモ"
	task.SetNoteID(noteID)

	if !task.HasNote() {
		t.Error("Task should have note after SetNoteID()")
	}

	if task.NoteID != noteID {
		t.Errorf("NoteID = %q, want %q", task.NoteID, noteID)
	}
}

func TestTaskHasNote(t *testing.T) {
	task := NewTask(1, "テスト", PriorityNone)

	if task.HasNote() {
		t.Error("HasNote() should return false for empty NoteID")
	}

	task.NoteID = "メモ"
	if !task.HasNote() {
		t.Error("HasNote() should return true for non-empty NoteID")
	}

	task.NoteID = ""
	if task.HasNote() {
		t.Error("HasNote() should return false after clearing NoteID")
	}
}

func TestCyclePriority(t *testing.T) {
	// Forward cycle: High -> Medium -> Low -> High
	got := CyclePriority(PriorityHigh, false)
	if got != PriorityMedium {
		t.Errorf("CyclePriority(High, false) = %v, want Medium", got)
	}

	got = CyclePriority(PriorityMedium, false)
	if got != PriorityLow {
		t.Errorf("CyclePriority(Medium, false) = %v, want Low", got)
	}

	got = CyclePriority(PriorityLow, false)
	if got != PriorityHigh {
		t.Errorf("CyclePriority(Low, false) = %v, want High", got)
	}

	// Reverse cycle: High -> Low -> Medium -> High
	got = CyclePriority(PriorityHigh, true)
	if got != PriorityLow {
		t.Errorf("CyclePriority(High, true) = %v, want Low", got)
	}

	got = CyclePriority(PriorityLow, true)
	if got != PriorityMedium {
		t.Errorf("CyclePriority(Low, true) = %v, want Medium", got)
	}

	// Unknown priority defaults to first in cycle
	got = CyclePriority(PriorityNone, false)
	if got != PriorityHigh {
		t.Errorf("CyclePriority(None, false) = %v, want High", got)
	}
}

func TestHasDueDate(t *testing.T) {
	task := NewTask(1, "テスト", PriorityMedium)

	if task.HasDueDate() {
		t.Error("New task should not have due date")
	}

	task.DueDate = time.Now().AddDate(0, 0, 1)
	if !task.HasDueDate() {
		t.Error("Task with DueDate set should have due date")
	}
}

func TestIsOverdue(t *testing.T) {
	task := NewTask(1, "テスト", PriorityMedium)

	// No due date -> not overdue
	if task.IsOverdue() {
		t.Error("Task without due date should not be overdue")
	}

	// Future due date -> not overdue
	task.DueDate = time.Now().AddDate(0, 0, 1)
	if task.IsOverdue() {
		t.Error("Task with future due date should not be overdue")
	}

	// Past due date -> overdue
	task.DueDate = time.Now().AddDate(0, 0, -1)
	if !task.IsOverdue() {
		t.Error("Task with past due date should be overdue")
	}

	// Done task -> not overdue
	task.Done()
	if task.IsOverdue() {
		t.Error("Done task should not be overdue")
	}
}

func TestIsDueSoon(t *testing.T) {
	task := NewTask(1, "テスト", PriorityMedium)

	// No due date -> not due soon
	if task.IsDueSoon(3) {
		t.Error("Task without due date should not be due soon")
	}

	// Due in 2 days with 3-day window -> due soon
	task.DueDate = time.Now().AddDate(0, 0, 2)
	if !task.IsDueSoon(3) {
		t.Error("Task due in 2 days should be due soon with 3-day window")
	}

	// Due in 5 days with 3-day window -> not due soon
	task.DueDate = time.Now().AddDate(0, 0, 5)
	if task.IsDueSoon(3) {
		t.Error("Task due in 5 days should not be due soon with 3-day window")
	}

	// Overdue -> not "due soon" (it's overdue)
	task.DueDate = time.Now().AddDate(0, 0, -1)
	if task.IsDueSoon(3) {
		t.Error("Overdue task should not be due soon")
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input string
		want  Priority
	}{
		{"1", PriorityHigh},
		{"high", PriorityHigh},
		{"2", PriorityMedium},
		{"medium", PriorityMedium},
		{"3", PriorityLow},
		{"low", PriorityLow},
		{"", PriorityNone},
		{"invalid", PriorityNone},
		{"4", PriorityNone},
	}

	for _, tt := range tests {
		got := ParsePriority(tt.input)
		if got != tt.want {
			t.Errorf("ParsePriority(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

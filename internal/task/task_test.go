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

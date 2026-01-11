package task

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "note-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	manager, err := NewManager(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create manager: %v", err)
	}

	return manager, tmpDir
}

func TestNewManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "note-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	notesDir := filepath.Join(tmpDir, "notes")
	manager, err := NewManager(notesDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		t.Error("NewManager() should create the directory")
	}
}

func TestManagerAdd(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テストタスク", PriorityHigh)

	if task == nil {
		t.Fatal("Add() returned nil")
	}

	if task.ID != 1 {
		t.Errorf("First task ID = %d, want 1", task.ID)
	}

	if task.Description != "テストタスク" {
		t.Errorf("Description = %q, want %q", task.Description, "テストタスク")
	}

	if task.Priority != PriorityHigh {
		t.Errorf("Priority = %d, want %d", task.Priority, PriorityHigh)
	}

	// Add second task
	task2 := manager.Add("タスク2", PriorityLow)
	if task2.ID != 2 {
		t.Errorf("Second task ID = %d, want 2", task2.ID)
	}
}

func TestManagerAddWithNote(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	noteID := "会議メモ"
	task := manager.AddWithNote("議事録まとめ", PriorityMedium, noteID)

	if task.NoteID != noteID {
		t.Errorf("NoteID = %q, want %q", task.NoteID, noteID)
	}

	if !task.HasNote() {
		t.Error("Task should have note after AddWithNote()")
	}
}

func TestManagerList(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	manager.Add("タスク1", PriorityLow)
	manager.Add("タスク2", PriorityHigh)
	task3 := manager.Add("タスク3", PriorityMedium)
	task3.Done()
	manager.save()

	// List without done
	list := manager.List(false)
	if len(list) != 2 {
		t.Errorf("List(false) returned %d tasks, want 2", len(list))
	}

	// List with done
	listAll := manager.List(true)
	if len(listAll) != 3 {
		t.Errorf("List(true) returned %d tasks, want 3", len(listAll))
	}

	// Check priority sorting (high first)
	if list[0].Priority != PriorityHigh {
		t.Error("List should be sorted by priority (high first)")
	}
}

func TestManagerListByNote(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	manager.AddWithNote("タスク1", PriorityHigh, "メモA")
	manager.AddWithNote("タスク2", PriorityMedium, "メモA")
	manager.AddWithNote("タスク3", PriorityLow, "メモB")
	manager.Add("タスク4", PriorityHigh)

	listA := manager.ListByNote("メモA")
	if len(listA) != 2 {
		t.Errorf("ListByNote(メモA) returned %d tasks, want 2", len(listA))
	}

	listB := manager.ListByNote("メモB")
	if len(listB) != 1 {
		t.Errorf("ListByNote(メモB) returned %d tasks, want 1", len(listB))
	}

	listC := manager.ListByNote("メモC")
	if len(listC) != 0 {
		t.Errorf("ListByNote(メモC) returned %d tasks, want 0", len(listC))
	}
}

func TestManagerGet(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テスト", PriorityMedium)

	got, err := manager.Get(task.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Description != task.Description {
		t.Errorf("Get().Description = %q, want %q", got.Description, task.Description)
	}

	_, err = manager.Get(999)
	if err == nil {
		t.Error("Get(non-existent) should return error")
	}
}

func TestManagerDone(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テスト", PriorityHigh)

	if task.IsDone() {
		t.Error("New task should not be done")
	}

	if err := manager.Done(task.ID); err != nil {
		t.Fatalf("Done() error = %v", err)
	}

	got, _ := manager.Get(task.ID)
	if !got.IsDone() {
		t.Error("Task should be done after Done()")
	}
}

func TestManagerDelete(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テスト", PriorityLow)

	if err := manager.Delete(task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := manager.Get(task.ID)
	if err == nil {
		t.Error("Get() after Delete() should return error")
	}

	if len(manager.List(true)) != 0 {
		t.Error("List should be empty after deleting the only task")
	}
}

func TestManagerToggle(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テスト", PriorityMedium)

	// Toggle to done
	if err := manager.Toggle(task.ID); err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	got, _ := manager.Get(task.ID)
	if !got.IsDone() {
		t.Error("Task should be done after first toggle")
	}

	// Toggle back to pending
	if err := manager.Toggle(task.ID); err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	got, _ = manager.Get(task.ID)
	if got.IsDone() {
		t.Error("Task should be pending after second toggle")
	}
}

func TestManagerSetNoteID(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.Add("テスト", PriorityHigh)

	if err := manager.SetNoteID(task.ID, "メモ"); err != nil {
		t.Fatalf("SetNoteID() error = %v", err)
	}

	got, _ := manager.Get(task.ID)
	if got.NoteID != "メモ" {
		t.Errorf("NoteID = %q, want %q", got.NoteID, "メモ")
	}
}

func TestManagerUnlinkNote(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	defer os.RemoveAll(tmpDir)

	task := manager.AddWithNote("テスト", PriorityMedium, "メモ")

	if err := manager.UnlinkNote(task.ID); err != nil {
		t.Fatalf("UnlinkNote() error = %v", err)
	}

	got, _ := manager.Get(task.ID)
	if got.HasNote() {
		t.Error("Task should not have note after UnlinkNote()")
	}
}

func TestManagerPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "note-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager and add tasks
	manager1, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	manager1.Add("タスク1", PriorityHigh)
	manager1.AddWithNote("タスク2", PriorityLow, "メモ")

	// Create new manager (simulating restart)
	manager2, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	list := manager2.List(true)
	if len(list) != 2 {
		t.Errorf("After reload, List() returned %d tasks, want 2", len(list))
	}

	task2, _ := manager2.Get(2)
	if task2.NoteID != "メモ" {
		t.Errorf("After reload, NoteID = %q, want %q", task2.NoteID, "メモ")
	}
}

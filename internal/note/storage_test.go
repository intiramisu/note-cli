package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "note-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	storage, err := NewStorage(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	return storage, tmpDir
}

func TestNewStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "note-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	notesDir := filepath.Join(tmpDir, "notes")
	storage, err := NewStorage(notesDir)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	if storage == nil {
		t.Fatal("NewStorage() returned nil")
	}

	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		t.Error("NewStorage() should create the directory")
	}
}

func TestStorageSaveAndLoad(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	note := NewNote("テストメモ", []string{"go", "test"})
	note.Content = "これはテストです"

	if err := storage.Save(note); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if note.ID == "" {
		t.Error("Save() should set note ID")
	}

	loaded, err := storage.Load(note.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Title != note.Title {
		t.Errorf("Loaded Title = %q, want %q", loaded.Title, note.Title)
	}

	if len(loaded.Tags) != len(note.Tags) {
		t.Errorf("Loaded Tags length = %d, want %d", len(loaded.Tags), len(note.Tags))
	}

	if !strings.Contains(loaded.Content, "テストメモ") {
		t.Error("Loaded Content should contain the title header")
	}
}

func TestStorageList(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	notes := []struct {
		title string
		tags  []string
	}{
		{"メモ1", []string{"tag1"}},
		{"メモ2", []string{"tag2"}},
		{"メモ3", []string{"tag1", "tag2"}},
	}

	for _, n := range notes {
		note := NewNote(n.title, n.tags)
		if err := storage.Save(note); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	list, err := storage.List("")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 3 {
		t.Errorf("List() returned %d notes, want 3", len(list))
	}

	// Test tag filter
	filtered, err := storage.List("tag1")
	if err != nil {
		t.Fatalf("List(tag1) error = %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("List(tag1) returned %d notes, want 2", len(filtered))
	}
}

func TestStorageFind(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	note := NewNote("検索テスト", []string{})
	if err := storage.Save(note); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Find by exact filename
	found, err := storage.Find(note.ID)
	if err != nil {
		t.Fatalf("Find(ID) error = %v", err)
	}
	if found.Title != note.Title {
		t.Errorf("Find(ID) Title = %q, want %q", found.Title, note.Title)
	}

	// Find by title
	found, err = storage.Find("検索テスト")
	if err != nil {
		t.Fatalf("Find(title) error = %v", err)
	}
	if found.Title != note.Title {
		t.Errorf("Find(title) Title = %q, want %q", found.Title, note.Title)
	}

	// Find by partial match
	found, err = storage.Find("検索")
	if err != nil {
		t.Fatalf("Find(partial) error = %v", err)
	}
	if found.Title != note.Title {
		t.Errorf("Find(partial) Title = %q, want %q", found.Title, note.Title)
	}

	// Find non-existent
	_, err = storage.Find("存在しないメモ")
	if err == nil {
		t.Error("Find(non-existent) should return error")
	}
}

func TestStorageDelete(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	note := NewNote("削除テスト", []string{})
	if err := storage.Save(note); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	filePath := storage.GetPath(note.ID)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("File should exist after save")
	}

	if err := storage.Delete(note.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should not exist after delete")
	}
}

func TestStorageGetPath(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	filename := "test.md"
	expected := filepath.Join(tmpDir, filename)

	got := storage.GetPath(filename)
	if got != expected {
		t.Errorf("GetPath() = %q, want %q", got, expected)
	}
}

func TestGenerateFilename(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		title    string
		expected string
	}{
		{"テスト", "テスト.md"},
		{"test memo", "test memo.md"},
		{"file/with:special*chars", "filewithspecialchars.md"},
		{"", "untitled.md"},
		{"   ", "untitled.md"},
	}

	for _, tt := range tests {
		note := NewNote(tt.title, nil)
		storage.Save(note)
		if note.ID != tt.expected {
			t.Errorf("generateFilename(%q) = %q, want %q", tt.title, note.ID, tt.expected)
		}
		storage.Delete(note.ID)
	}
}

func TestParseNoteNoFrontmatter(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	// Write a file without frontmatter
	content := "Just some text without frontmatter"
	filename := "no-frontmatter.md"
	fullPath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := storage.Load(filename)
	if err == nil {
		t.Error("Load() should return error for file without frontmatter")
	}
	if !strings.Contains(err.Error(), "frontmatter") {
		t.Errorf("Error should mention frontmatter, got: %v", err)
	}
}

func TestParseNoteNoClosingFrontmatter(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	// Write a file with unclosed frontmatter
	content := "---\ntitle: test\n"
	filename := "unclosed.md"
	fullPath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := storage.Load(filename)
	if err == nil {
		t.Error("Load() should return error for unclosed frontmatter")
	}
	if !strings.Contains(err.Error(), "終端") {
		t.Errorf("Error should mention missing end marker, got: %v", err)
	}
}

func TestParseNoteInvalidYAML(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	// Write a file with invalid YAML in frontmatter
	content := "---\ntitle: [invalid yaml\n---\ncontent"
	filename := "invalid-yaml.md"
	fullPath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := storage.Load(filename)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestStorageSaveAt(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	n := NewNote("SaveAtテスト", []string{"test"})
	n.Content = "テスト内容"

	customPath := filepath.Join(tmpDir, "subdir", "custom.md")
	os.MkdirAll(filepath.Dir(customPath), 0755)

	if err := storage.SaveAt(n, customPath); err != nil {
		t.Fatalf("SaveAt() error = %v", err)
	}

	// Verify file was created at custom path
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Error("SaveAt() should create file at specified path")
	}

	// Verify content
	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "SaveAtテスト") {
		t.Error("Saved file should contain the title")
	}
	if !strings.Contains(content, "テスト内容") {
		t.Error("Saved file should contain the content")
	}
}

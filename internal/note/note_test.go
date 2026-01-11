package note

import (
	"testing"
	"time"
)

func TestNewNote(t *testing.T) {
	title := "テストメモ"
	tags := []string{"go", "test"}

	before := time.Now()
	n := NewNote(title, tags)
	after := time.Now()

	if n.Title != title {
		t.Errorf("Title = %q, want %q", n.Title, title)
	}

	if len(n.Tags) != len(tags) {
		t.Errorf("Tags length = %d, want %d", len(n.Tags), len(tags))
	}

	for i, tag := range tags {
		if n.Tags[i] != tag {
			t.Errorf("Tags[%d] = %q, want %q", i, n.Tags[i], tag)
		}
	}

	if n.Created.Before(before) || n.Created.After(after) {
		t.Errorf("Created time out of range")
	}

	if n.Modified.Before(before) || n.Modified.After(after) {
		t.Errorf("Modified time out of range")
	}

	if n.Content != "" {
		t.Errorf("Content = %q, want empty", n.Content)
	}
}

func TestNewNoteEmptyTags(t *testing.T) {
	n := NewNote("タイトル", []string{})

	if n.Tags == nil {
		t.Error("Tags should not be nil")
	}

	if len(n.Tags) != 0 {
		t.Errorf("Tags length = %d, want 0", len(n.Tags))
	}
}

func TestNewNoteNilTags(t *testing.T) {
	n := NewNote("タイトル", nil)

	if n.Tags != nil {
		t.Error("Tags should be nil when passed nil")
	}
}

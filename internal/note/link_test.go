package note

import (
	"os"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			"single link",
			"See [[メモA]] for details",
			[]string{"メモA"},
		},
		{
			"multiple links",
			"See [[メモA]] and [[メモB]]",
			[]string{"メモA", "メモB"},
		},
		{
			"duplicate links",
			"See [[メモA]] and [[メモA]] again",
			[]string{"メモA"},
		},
		{
			"no links",
			"No links here",
			nil,
		},
		{
			"empty brackets",
			"See [[]] nothing",
			nil,
		},
		{
			"link with spaces",
			"See [[ メモ C ]] padded",
			[]string{"メモ C"},
		},
		{
			"link in code block context",
			"See `[[not a link]]` but [[メモD]] is",
			[]string{"not a link", "メモD"},
		},
		{
			"multiline",
			"Line1 [[メモA]]\nLine2 [[メモB]]",
			[]string{"メモA", "メモB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractLinks(tt.content)
			if len(got) != len(tt.expected) {
				t.Errorf("ExtractLinks() returned %d links, want %d: got=%v", len(got), len(tt.expected), got)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("ExtractLinks()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestResolveLinks(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	// Create test notes
	noteA := NewNote("メモA", nil)
	noteA.Content = "Content A"
	storage.Save(noteA)

	noteB := NewNote("メモB", nil)
	noteB.Content = "Content B"
	storage.Save(noteB)

	links := []string{"メモA", "メモB", "存在しない"}
	found, notFound := ResolveLinks(storage, links)

	if len(found) != 2 {
		t.Errorf("ResolveLinks() found %d notes, want 2", len(found))
	}
	if len(notFound) != 1 {
		t.Errorf("ResolveLinks() notFound %d, want 1", len(notFound))
	}
	if len(notFound) > 0 && notFound[0] != "存在しない" {
		t.Errorf("ResolveLinks() notFound[0] = %q, want %q", notFound[0], "存在しない")
	}
}

func TestFindBacklinks(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	// Create notes with links
	noteA := NewNote("メモA", nil)
	noteA.Content = "This links to [[メモC]]"
	storage.Save(noteA)

	noteB := NewNote("メモB", nil)
	noteB.Content = "This also links to [[メモC]]"
	storage.Save(noteB)

	noteC := NewNote("メモC", nil)
	noteC.Content = "Target note"
	storage.Save(noteC)

	backlinks, err := FindBacklinks(storage, "メモC")
	if err != nil {
		t.Fatalf("FindBacklinks() error = %v", err)
	}

	if len(backlinks) != 2 {
		t.Errorf("FindBacklinks() returned %d backlinks, want 2", len(backlinks))
	}

	// メモC自身はバックリンクに含まれないこと
	for _, bl := range backlinks {
		if bl.Title == "メモC" {
			t.Error("FindBacklinks() should not include the target note itself")
		}
	}
}

func TestFindBacklinksNoResults(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	noteA := NewNote("メモA", nil)
	noteA.Content = "No links here"
	storage.Save(noteA)

	backlinks, err := FindBacklinks(storage, "メモA")
	if err != nil {
		t.Fatalf("FindBacklinks() error = %v", err)
	}

	if len(backlinks) != 0 {
		t.Errorf("FindBacklinks() returned %d backlinks, want 0", len(backlinks))
	}
}

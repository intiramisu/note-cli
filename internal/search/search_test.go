package search

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "note-cli-search-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func writeTestNote(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
}

func TestSearchBasic(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "test1.md", `---
title: テストメモ1
---
これはテストです
検索対象の文字列`)

	writeTestNote(t, tmpDir, "test2.md", `---
title: テストメモ2
---
別のメモです
検索対象の文字列がここにも`)

	results, err := Search(tmpDir, "検索対象")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search() returned %d results, want 2", len(results))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "test.md", `---
title: Test Note
---
Hello World
HELLO again
hello there`)

	results, err := Search(tmpDir, "hello")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Search(hello) returned %d results, want 3", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "test.md", `---
title: メモ
---
内容です`)

	results, err := Search(tmpDir, "存在しない文字列")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Search() returned %d results, want 0", len(results))
	}
}

func TestSearchLineNumbers(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "test.md", `---
title: メモ
---
行1
行2
検索対象
行4`)

	results, err := Search(tmpDir, "検索対象")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}

	if results[0].Line != 6 {
		t.Errorf("Line = %d, want 6", results[0].Line)
	}
}

func TestSearchTitle(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "memo.md", `---
title: 会議メモ
---
内容`)

	results, err := Search(tmpDir, "内容")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}

	if results[0].Title != "会議メモ" {
		t.Errorf("Title = %q, want %q", results[0].Title, "会議メモ")
	}
}

func TestSearchSkipsNonMarkdown(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "test.md", `---
title: Markdown
---
検索対象`)

	// Non-markdown file
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("検索対象"), 0644); err != nil {
		t.Fatalf("Failed to write txt file: %v", err)
	}

	results, err := Search(tmpDir, "検索対象")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search() returned %d results, want 1 (should skip .txt)", len(results))
	}
}

func TestSearchSkipsDirectories(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "root.md", `---
title: Root
---
検索対象`)

	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	writeTestNote(t, subDir, "sub.md", `---
title: Sub
---
検索対象`)

	results, err := Search(tmpDir, "検索対象")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	// Should find both files (walks subdirectories)
	if len(results) != 2 {
		t.Errorf("Search() returned %d results, want 2", len(results))
	}
}

func TestSearchEmptyDir(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	results, err := Search(tmpDir, "anything")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Search() returned %d results, want 0", len(results))
	}
}

func TestResultStruct(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	writeTestNote(t, tmpDir, "note.md", `---
title: タイトル
---
マッチする行`)

	results, err := Search(tmpDir, "マッチする行")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}

	r := results[0]
	if r.Filename != "note.md" {
		t.Errorf("Filename = %q, want %q", r.Filename, "note.md")
	}

	if r.Title != "タイトル" {
		t.Errorf("Title = %q, want %q", r.Title, "タイトル")
	}

	if r.Content != "マッチする行" {
		t.Errorf("Content = %q, want %q", r.Content, "マッチする行")
	}
}

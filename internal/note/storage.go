package note

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Storage struct {
	notesDir string
}

func NewStorage(notesDir string) (*Storage, error) {
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("メモディレクトリの作成に失敗: %w", err)
	}
	return &Storage{notesDir: notesDir}, nil
}

func (s *Storage) Save(note *Note) error {
	filename := s.generateFilename(note.Title)
	fullPath := filepath.Join(s.notesDir, filename)

	content := s.formatNote(note)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("メモの保存に失敗: %w", err)
	}

	note.ID = filename
	return nil
}

func (s *Storage) SaveAt(note *Note, path string) error {
	content := s.formatNote(note)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("メモの保存に失敗: %w", err)
	}
	return nil
}

func (s *Storage) List(tagFilter string) ([]*Note, error) {
	var notes []*Note

	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // エラーがあってもスキップして続行
		}

		// .templates ディレクトリはスキップ
		if d.IsDir() && d.Name() == ".templates" {
			return filepath.SkipDir
		}

		// ディレクトリや .md 以外はスキップ
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// notesDir からの相対パスを取得
		relPath, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return nil
		}

		note, err := s.Load(relPath)
		if err != nil {
			return nil
		}

		if tagFilter != "" {
			hasTag := false
			for _, tag := range note.Tags {
				if tag == tagFilter {
					hasTag = true
					break
				}
			}
			if !hasTag {
				return nil
			}
		}

		notes = append(notes, note)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("メモディレクトリの読み取りに失敗: %w", err)
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Modified.After(notes[j].Modified)
	})

	return notes, nil
}

func (s *Storage) Load(filename string) (*Note, error) {
	fullPath := filepath.Join(s.notesDir, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("メモの読み込みに失敗: %w", err)
	}

	return s.parseNote(filename, string(data))
}

func (s *Storage) generateFilename(title string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	safe := reg.ReplaceAllString(title, "")
	safe = strings.TrimSpace(safe)
	if safe == "" {
		safe = "untitled"
	}
	return safe + ".md"
}

func (s *Storage) formatNote(note *Note) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", note.Title))
	sb.WriteString(fmt.Sprintf("created: %s\n", note.Created.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("modified: %s\n", note.Modified.Format(time.RFC3339)))
	if len(note.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(note.Tags, ", ")))
	} else {
		sb.WriteString("tags: []\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", note.Title))
	sb.WriteString(note.Content)

	return sb.String()
}

func (s *Storage) Find(query string) (*Note, error) {
	if strings.HasSuffix(query, ".md") {
		if note, err := s.Load(query); err == nil {
			return note, nil
		}
	}

	if note, err := s.Load(query + ".md"); err == nil {
		return note, nil
	}

	notes, err := s.List("")
	if err != nil {
		return nil, err
	}

	for _, n := range notes {
		if strings.EqualFold(n.Title, query) {
			return n, nil
		}
	}

	for _, n := range notes {
		if strings.Contains(strings.ToLower(n.Title), strings.ToLower(query)) {
			return n, nil
		}
	}

	return nil, fmt.Errorf("メモが見つかりません: %s", query)
}

func (s *Storage) Delete(filename string) error {
	path := filepath.Join(s.notesDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("メモの削除に失敗: %w", err)
	}
	return nil
}

func (s *Storage) GetPath(filename string) string {
	return filepath.Join(s.notesDir, filename)
}

func (s *Storage) parseNote(filename, content string) (*Note, error) {
	note := &Note{ID: filename}

	// frontmatterを抽出
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("frontmatterが見つかりません")
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("frontmatterの終端が見つかりません")
	}

	// frontmatterをパース
	var frontmatter struct {
		Title    string    `yaml:"title"`
		Created  time.Time `yaml:"created"`
		Modified time.Time `yaml:"modified"`
		Tags     []string  `yaml:"tags"`
	}

	if err := yaml.Unmarshal([]byte(parts[0]), &frontmatter); err != nil {
		return nil, fmt.Errorf("frontmatterのパースに失敗: %w", err)
	}

	note.Title = frontmatter.Title
	note.Created = frontmatter.Created
	note.Modified = frontmatter.Modified
	note.Tags = frontmatter.Tags
	note.Content = strings.TrimSpace(parts[1])

	return note, nil
}

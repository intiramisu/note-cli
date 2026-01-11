package search

import (
	"os"
	"path/filepath"
	"strings"
)

type Result struct {
	Filename string
	Title    string
	Line     int
	Content  string
}

func Search(notesDir, query string) ([]Result, error) {
	var results []Result
	queryLower := strings.ToLower(query)

	err := filepath.Walk(notesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")

		title := info.Name()
		for _, line := range lines {
			if strings.HasPrefix(line, "title:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				break
			}
		}

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				results = append(results, Result{
					Filename: info.Name(),
					Title:    title,
					Line:     i + 1,
					Content:  strings.TrimSpace(line),
				})
			}
		}

		return nil
	})

	return results, err
}

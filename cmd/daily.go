package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/spf13/cobra"
)

var dailyCmd = &cobra.Command{
	Use:     "daily [date]",
	Aliases: []string{"d"},
	Short:   "Open daily note",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Global
		date := time.Now()
		if len(args) > 0 {
			parsed, err := parseDate(args[0], cfg.Formats.Date)
			if err != nil {
				return err
			}
			date = parsed
		}

		notesDir := cfg.NotesDir
		storage, err := note.NewStorage(notesDir)
		if err != nil {
			return err
		}

		// daily ディレクトリを確保
		dailyDir := filepath.Join(notesDir, cfg.Paths.DailyDir)
		if err := os.MkdirAll(dailyDir, 0755); err != nil {
			return fmt.Errorf("dailyディレクトリの作成に失敗: %w", err)
		}

		dateStr := date.Format(cfg.Formats.Date)
		filename := dateStr + ".md"
		filePath := filepath.Join(dailyDir, filename)

		// 既存のノートがあれば開く
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("%s %s を開きます\n", cfg.Theme.Symbols.DailyIcon, dateStr)
			return openEditor(filePath)
		}

		// 新規作成
		content, err := loadDailyTemplate(notesDir, date, cfg)
		if err != nil {
			return err
		}

		n := &note.Note{
			ID:       filepath.Join(cfg.Paths.DailyDir, dateStr),
			Title:    dateStr,
			Created:  time.Now(),
			Modified: time.Now(),
			Tags:     []string{"daily"},
			Content:  content,
		}

		if err := storage.SaveAt(n, filePath); err != nil {
			return err
		}

		fmt.Printf("%s %s を作成しました\n", cfg.Theme.Symbols.DailyIcon, dateStr)
		return openEditor(filePath)
	},
}

func parseDate(input string, dateFormat string) (time.Time, error) {
	now := time.Now()

	switch strings.ToLower(input) {
	case "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	}

	// +N / -N 形式
	if len(input) > 0 && (input[0] == '+' || input[0] == '-') {
		var days int
		if _, err := fmt.Sscanf(input, "%d", &days); err == nil {
			return now.AddDate(0, 0, days), nil
		}
	}

	// 日付形式 (設定から取得)
	parsed, err := time.Parse(dateFormat, input)
	if err != nil {
		return time.Time{}, fmt.Errorf("無効な日付形式: %s (%s, yesterday, tomorrow, +N, -N が使えます)", input, dateFormat)
	}
	return parsed, nil
}

func loadDailyTemplate(notesDir string, date time.Time, cfg *config.Config) (string, error) {
	templatePath := filepath.Join(notesDir, cfg.Paths.TemplatesDir, "daily.md")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		// テンプレートがなければデフォルト
		return getDefaultDailyContent(date, cfg), nil
	}

	// テンプレート内の変数を置換
	content := string(data)
	content = strings.ReplaceAll(content, "{{date}}", date.Format(cfg.Formats.Date))
	content = strings.ReplaceAll(content, "{{year}}", date.Format("2006"))
	content = strings.ReplaceAll(content, "{{month}}", date.Format("01"))
	content = strings.ReplaceAll(content, "{{day}}", date.Format("02"))
	content = strings.ReplaceAll(content, "{{weekday}}", date.Weekday().String())

	return content, nil
}

func getDefaultDailyContent(date time.Time, cfg *config.Config) string {
	dateStr := date.Format(cfg.Formats.Date)
	weekday := getJapaneseWeekday(date.Weekday())

	return fmt.Sprintf(`## やること

- [ ]

## メモ

## 振り返り

---
%s (%s)
`, dateStr, weekday)
}

func getJapaneseWeekday(w time.Weekday) string {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return weekdays[w]
}

func init() {
	rootCmd.AddCommand(dailyCmd)
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/intiramisu/note-cli/internal/note"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dailyCmd = &cobra.Command{
	Use:     "daily [日付]",
	Aliases: []string{"d"},
	Short:   "デイリーノートを開く",
	Long: `今日のデイリーノートを開きます。存在しない場合は新規作成します。

日付の指定方法:
  note-cli daily              # 今日
  note-cli daily yesterday    # 昨日
  note-cli daily tomorrow     # 明日
  note-cli daily 2025-01-11   # 指定日 (YYYY-MM-DD)
  note-cli daily -1           # 1日前
  note-cli daily +1           # 1日後`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		date := time.Now()
		if len(args) > 0 {
			parsed, err := parseDate(args[0])
			if err != nil {
				return err
			}
			date = parsed
		}

		notesDir := viper.GetString("notes_dir")
		storage, err := note.NewStorage(notesDir)
		if err != nil {
			return err
		}

		// daily ディレクトリを確保
		dailyDir := filepath.Join(notesDir, "daily")
		if err := os.MkdirAll(dailyDir, 0755); err != nil {
			return fmt.Errorf("dailyディレクトリの作成に失敗: %w", err)
		}

		dateStr := date.Format("2006-01-02")
		filename := dateStr + ".md"
		filePath := filepath.Join(dailyDir, filename)

		// 既存のノートがあれば開く
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("📅 %s を開きます\n", dateStr)
			return openEditor(filePath)
		}

		// 新規作成
		content, err := loadDailyTemplate(notesDir, date)
		if err != nil {
			return err
		}

		n := &note.Note{
			ID:       filepath.Join("daily", dateStr),
			Title:    dateStr,
			Created:  time.Now(),
			Modified: time.Now(),
			Tags:     []string{"daily"},
			Content:  content,
		}

		if err := storage.SaveAt(n, filePath); err != nil {
			return err
		}

		fmt.Printf("📅 %s を作成しました\n", dateStr)
		return openEditor(filePath)
	},
}

func parseDate(input string) (time.Time, error) {
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

	// YYYY-MM-DD 形式
	parsed, err := time.Parse("2006-01-02", input)
	if err != nil {
		return time.Time{}, fmt.Errorf("無効な日付形式: %s (YYYY-MM-DD, yesterday, tomorrow, +N, -N が使えます)", input)
	}
	return parsed, nil
}

func loadDailyTemplate(notesDir string, date time.Time) (string, error) {
	templatePath := filepath.Join(notesDir, ".templates", "daily.md")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		// テンプレートがなければデフォルト
		return getDefaultDailyContent(date), nil
	}

	// テンプレート内の変数を置換
	content := string(data)
	content = strings.ReplaceAll(content, "{{date}}", date.Format("2006-01-02"))
	content = strings.ReplaceAll(content, "{{year}}", date.Format("2006"))
	content = strings.ReplaceAll(content, "{{month}}", date.Format("01"))
	content = strings.ReplaceAll(content, "{{day}}", date.Format("02"))
	content = strings.ReplaceAll(content, "{{weekday}}", date.Weekday().String())

	return content, nil
}

func getDefaultDailyContent(date time.Time) string {
	dateStr := date.Format("2006-01-02")
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

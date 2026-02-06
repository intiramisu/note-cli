package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config はアプリケーション全体の設定を保持する
type Config struct {
	NotesDir    string   `mapstructure:"notes_dir"`
	Editor      string   `mapstructure:"editor"`
	DefaultTags []string `mapstructure:"default_tags"`
	Paths       Paths    `mapstructure:"paths"`
	Formats     Formats  `mapstructure:"formats"`
	Theme       Theme    `mapstructure:"theme"`
	Display     Display  `mapstructure:"display"`
}

// Paths はパス関連の設定
type Paths struct {
	TemplatesDir string `mapstructure:"templates_dir"`
	TasksFile    string `mapstructure:"tasks_file"`
	DailyDir     string `mapstructure:"daily_dir"`
}

// Formats は日付フォーマットの設定
type Formats struct {
	Date     string `mapstructure:"date"`
	DateTime string `mapstructure:"datetime"`
}

// Theme はテーマ設定
type Theme struct {
	Colors   Colors   `mapstructure:"colors"`
	Symbols  Symbols  `mapstructure:"symbols"`
	Sections Sections `mapstructure:"sections"`
}

// Colors はカラー設定 (hex or 256色)
type Colors struct {
	Title          string `mapstructure:"title"`
	Selected       string `mapstructure:"selected"`
	Done           string `mapstructure:"done"`
	Help           string `mapstructure:"help"`
	Empty          string `mapstructure:"empty"`
	PriorityHigh   string `mapstructure:"priority_high"`
	PriorityMedium string `mapstructure:"priority_medium"`
	PriorityLow    string `mapstructure:"priority_low"`
}

// Symbols はシンボル設定
type Symbols struct {
	Cursor        string `mapstructure:"cursor"`
	CursorEmpty   string `mapstructure:"cursor_empty"`
	CheckboxEmpty string `mapstructure:"checkbox_empty"`
	CheckboxDone  string `mapstructure:"checkbox_done"`
	NoteIcon      string `mapstructure:"note_icon"`
	TaskIcon      string `mapstructure:"task_icon"`
	DailyIcon     string `mapstructure:"daily_icon"`
}

// Sections はセクション名設定
type Sections struct {
	P1   string `mapstructure:"p1"`
	P2   string `mapstructure:"p2"`
	P3   string `mapstructure:"p3"`
	Done string `mapstructure:"done"`
}

// Display は表示設定
type Display struct {
	SeparatorWidth int `mapstructure:"separator_width"`
	TaskCharLimit  int `mapstructure:"task_char_limit"`
	InputWidth     int `mapstructure:"input_width"`
}

// Global は現在の設定を保持するグローバル変数
var Global *Config

// SetDefaults はデフォルト値を設定する
func SetDefaults() {
	home, _ := os.UserHomeDir()

	// 基本設定
	viper.SetDefault("notes_dir", filepath.Join(home, "notes"))
	viper.SetDefault("editor", "vim")
	viper.SetDefault("default_tags", []string{})

	// パス設定
	viper.SetDefault("paths.templates_dir", ".templates")
	viper.SetDefault("paths.tasks_file", ".tasks.yaml")
	viper.SetDefault("paths.daily_dir", "daily")

	// フォーマット設定
	viper.SetDefault("formats.date", "2006-01-02")
	viper.SetDefault("formats.datetime", "2006-01-02 15:04")

	// テーマ - カラー
	viper.SetDefault("theme.colors.title", "#cd7cf4")
	viper.SetDefault("theme.colors.selected", "#d75fd7")
	viper.SetDefault("theme.colors.done", "#626262")
	viper.SetDefault("theme.colors.help", "#626262")
	viper.SetDefault("theme.colors.empty", "#585858")
	viper.SetDefault("theme.colors.priority_high", "#ff0000")
	viper.SetDefault("theme.colors.priority_medium", "#ffaf00")
	viper.SetDefault("theme.colors.priority_low", "#5fafff")

	// テーマ - シンボル
	viper.SetDefault("theme.symbols.cursor", "▸ ")
	viper.SetDefault("theme.symbols.cursor_empty", "  ")
	viper.SetDefault("theme.symbols.checkbox_empty", "[ ]")
	viper.SetDefault("theme.symbols.checkbox_done", "[✓]")
	viper.SetDefault("theme.symbols.note_icon", "📄")
	viper.SetDefault("theme.symbols.task_icon", "📋")
	viper.SetDefault("theme.symbols.daily_icon", "📅")

	// テーマ - セクション名
	viper.SetDefault("theme.sections.p1", "🔥 P1")
	viper.SetDefault("theme.sections.p2", "⚡ P2")
	viper.SetDefault("theme.sections.p3", "📝 P3")
	viper.SetDefault("theme.sections.done", "✅ 完了")

	// 表示設定
	viper.SetDefault("display.separator_width", 40)
	viper.SetDefault("display.task_char_limit", 100)
	viper.SetDefault("display.input_width", 40)
}

// Load は設定を読み込んでグローバル変数に格納する
func Load() error {
	Global = &Config{}
	if err := viper.Unmarshal(Global); err != nil {
		return err
	}
	// ~/ を展開
	Global.NotesDir = expandTilde(Global.NotesDir)
	return nil
}

// expandTilde はパスの先頭の ~/ をホームディレクトリに展開する
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// GetTemplatesPath はテンプレートディレクトリの絶対パスを返す
func (c *Config) GetTemplatesPath() string {
	return filepath.Join(c.NotesDir, c.Paths.TemplatesDir)
}

// GetTasksPath はタスクファイルの絶対パスを返す
func (c *Config) GetTasksPath() string {
	return filepath.Join(c.NotesDir, c.Paths.TasksFile)
}

// GetDailyPath はデイリーノートディレクトリの絶対パスを返す
func (c *Config) GetDailyPath() string {
	return filepath.Join(c.NotesDir, c.Paths.DailyDir)
}

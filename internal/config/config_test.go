package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	viper.Reset()

	// Use a fresh viper for this test
	_ = v

	SetDefaults()

	// Check basic defaults
	notesDir := viper.GetString("notes_dir")
	if notesDir == "" {
		t.Error("notes_dir should have a default value")
	}
	home, _ := os.UserHomeDir()
	if !strings.Contains(notesDir, home) {
		t.Errorf("notes_dir should contain home directory, got %q", notesDir)
	}

	editor := viper.GetString("editor")
	if editor != "vim" {
		t.Errorf("editor = %q, want %q", editor, "vim")
	}

	// Check path defaults
	tasksFile := viper.GetString("paths.tasks_file")
	if tasksFile != ".tasks.yaml" {
		t.Errorf("paths.tasks_file = %q, want %q", tasksFile, ".tasks.yaml")
	}

	dailyDir := viper.GetString("paths.daily_dir")
	if dailyDir != "daily" {
		t.Errorf("paths.daily_dir = %q, want %q", dailyDir, "daily")
	}

	// Check format defaults
	dateFormat := viper.GetString("formats.date")
	if dateFormat != "2006-01-02" {
		t.Errorf("formats.date = %q, want %q", dateFormat, "2006-01-02")
	}

	// Check theme defaults
	titleColor := viper.GetString("theme.colors.title")
	if titleColor == "" {
		t.Error("theme.colors.title should have a default value")
	}
}

func TestLoad(t *testing.T) {
	viper.Reset()
	SetDefaults()

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if Global == nil {
		t.Fatal("Global should be set after Load()")
	}

	if Global.Editor != "vim" {
		t.Errorf("Global.Editor = %q, want %q", Global.Editor, "vim")
	}

	if Global.Paths.TasksFile != ".tasks.yaml" {
		t.Errorf("Global.Paths.TasksFile = %q, want %q", Global.Paths.TasksFile, ".tasks.yaml")
	}

	// NotesDir should be expanded (no ~/)
	if strings.HasPrefix(Global.NotesDir, "~/") {
		t.Errorf("NotesDir should be expanded, got %q", Global.NotesDir)
	}
}

func TestExpandTilde(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "path with tilde",
			path: "~/notes",
			want: home + "/notes",
		},
		{
			name: "absolute path",
			path: "/tmp/notes",
			want: "/tmp/notes",
		},
		{
			name: "relative path",
			path: "notes",
			want: "notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTilde(tt.path)
			if got != tt.want {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/search"
	"github.com/spf13/cobra"
)

// ルートレベルのショートカットコマンド（メモ操作をより短く）

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		templateName, _ := cmd.Flags().GetString("template")

		notesDir := config.Global.NotesDir
		storage, err := note.NewStorage(notesDir)
		if err != nil {
			return err
		}

		n := note.NewNote(title, tags)

		// テンプレートがあれば読み込み
		if templateName != "" {
			content, err := loadTemplate(notesDir, templateName, title)
			if err != nil {
				return err
			}
			n.Content = content
		}

		if err := storage.Save(n); err != nil {
			return err
		}

		fmt.Printf("メモを作成しました: %s\n", n.ID)
		return openEditor(storage.GetPath(n.ID))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		tagFilter, _ := cmd.Flags().GetString("tag")
		cfg := config.Global

		storage, err := note.NewStorage(cfg.NotesDir)
		if err != nil {
			return err
		}

		notes, err := storage.List(tagFilter)
		if err != nil {
			return err
		}

		if len(notes) == 0 {
			fmt.Println("メモがありません")
			return nil
		}

		for _, n := range notes {
			tagsStr := ""
			if len(n.Tags) > 0 {
				tagsStr = " [" + strings.Join(n.Tags, ", ") + "]"
			}
			// サブディレクトリにあるノートはパスを表示
			dir := filepath.Dir(n.ID)
			titleDisplay := n.Title
			if dir != "." {
				titleDisplay = dir + "/" + n.Title
			}
			fmt.Printf("- %s%s (%s)\n", titleDisplay, tagsStr, n.Modified.Format(cfg.Formats.DateTime))
		}

		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <title>",
	Short: "Show note content",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		cfg := config.Global

		storage, err := note.NewStorage(cfg.NotesDir)
		if err != nil {
			return err
		}

		n, err := storage.Find(query)
		if err != nil {
			return err
		}

		fmt.Printf("# %s\n", n.Title)
		fmt.Printf("作成: %s | 更新: %s\n", n.Created.Format(cfg.Formats.DateTime), n.Modified.Format(cfg.Formats.DateTime))
		if len(n.Tags) > 0 {
			fmt.Printf("タグ: %s\n", strings.Join(n.Tags, ", "))
		}
		fmt.Println(strings.Repeat("-", cfg.Display.SeparatorWidth))
		fmt.Println(n.Content)

		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit <title>",
	Short: "Edit a note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		storage, err := note.NewStorage(config.Global.NotesDir)
		if err != nil {
			return err
		}

		n, err := storage.Find(query)
		if err != nil {
			return err
		}

		return openEditor(storage.GetPath(n.ID))
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		cfg := config.Global

		results, err := search.Search(cfg.NotesDir, query)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Printf("「%s」に一致するメモはありません\n", query)
			return nil
		}

		fmt.Printf("「%s」の検索結果: %d件\n\n", query, len(results))

		truncateWidth := cfg.Display.SearchTruncate
		currentFile := ""
		for _, r := range results {
			if r.Filename != currentFile {
				fmt.Printf("%s %s\n", cfg.Theme.Symbols.NoteIcon, r.Title)
				currentFile = r.Filename
			}
			content := r.Content
			if len(content) > truncateWidth {
				content = content[:truncateWidth-3] + "..."
			}
			fmt.Printf("   L%d: %s\n", r.Line, content)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(searchCmd)

	createCmd.Flags().StringSliceP("tag", "t", []string{}, "tags (can be specified multiple times)")
	createCmd.Flags().StringP("template", "T", "", "template name")
	listCmd.Flags().StringP("tag", "t", "", "filter by tag")
}

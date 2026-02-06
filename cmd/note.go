package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func openEditor(filePath string) error {
	editor := viper.GetString("editor")
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var noteCmd = &cobra.Command{
	Use:     "note",
	Aliases: []string{"n"},
	Short:   "Manage notes",
}

var noteCreateCmd = &cobra.Command{
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

func loadTemplate(notesDir, name, title string) (string, error) {
	cfg := config.Global
	templatesDir := cfg.Paths.TemplatesDir
	templatePath := filepath.Join(notesDir, templatesDir, name+".md")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("テンプレートが見つかりません: %s", name)
	}

	content := string(data)
	content = strings.ReplaceAll(content, "{{title}}", title)
	return content, nil
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		tagFilter, _ := cmd.Flags().GetString("tag")
		cfg := config.Global

		storage, err := newStorage()
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

var noteShowCmd = &cobra.Command{
	Use:   "show <title>",
	Short: "Show note content",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		cfg := config.Global

		storage, err := newStorage()
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

		// リンク情報を表示
		links := note.ExtractLinks(n.Content)
		if len(links) > 0 {
			fmt.Println()
			fmt.Println("🔗 リンク先:")
			found, notFound := note.ResolveLinks(storage, links)
			for _, ln := range found {
				fmt.Printf("  ✓ %s\n", ln.Title)
			}
			for _, name := range notFound {
				fmt.Printf("  ✗ %s (未作成)\n", name)
			}
		}

		// バックリンク情報を表示
		backlinks, err := note.FindBacklinks(storage, n.Title)
		if err == nil && len(backlinks) > 0 {
			fmt.Println()
			fmt.Println("🔙 被参照:")
			for _, bl := range backlinks {
				fmt.Printf("  ← %s\n", bl.Title)
			}
		}

		return nil
	},
}

var noteEditCmd = &cobra.Command{
	Use:   "edit <title>",
	Short: "Edit a note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		storage, err := newStorage()
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

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <title>",
	Short: "Delete a note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		force, _ := cmd.Flags().GetBool("force")

		storage, err := newStorage()
		if err != nil {
			return err
		}

		n, err := storage.Find(query)
		if err != nil {
			return err
		}

		if !force {
			fmt.Printf("メモ「%s」を削除しますか？ [y/N]: ", n.Title)
			var answer string
			fmt.Scanln(&answer)
			if strings.ToLower(answer) != "y" {
				fmt.Println("キャンセルしました")
				return nil
			}
		}

		if err := storage.Delete(n.ID); err != nil {
			return err
		}

		fmt.Printf("メモ「%s」を削除しました\n", n.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteShowCmd)
	noteCmd.AddCommand(noteEditCmd)
	noteCmd.AddCommand(noteDeleteCmd)

	noteCreateCmd.Flags().StringSliceP("tag", "t", []string{}, "tags (can be specified multiple times)")
	noteCreateCmd.Flags().StringP("template", "T", "", "template name")
	noteListCmd.Flags().StringP("tag", "t", "", "filter by tag")
	noteDeleteCmd.Flags().BoolP("force", "f", false, "delete without confirmation")
}

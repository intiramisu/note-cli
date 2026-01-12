package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/search"
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
	Short:   "メモの操作",
	Long:    `メモの作成、編集、一覧表示、検索などを行います。`,
}

var noteCreateCmd = &cobra.Command{
	Use:   "create <タイトル>",
	Short: "新規メモを作成",
	Long:  `指定したタイトルで新しいメモを作成し、エディタで開きます。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		templateName, _ := cmd.Flags().GetString("template")

		notesDir := viper.GetString("notes_dir")
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
	Short: "メモの一覧を表示",
	Long:  `保存されているメモの一覧を表示します。`,
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
			fmt.Printf("- %s%s (%s)\n", n.Title, tagsStr, n.Modified.Format(cfg.Formats.DateTime))
		}

		return nil
	},
}

var noteShowCmd = &cobra.Command{
	Use:   "show <タイトル|ファイル名>",
	Short: "メモの内容を表示",
	Long:  `指定したメモの内容を表示します。`,
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

var noteEditCmd = &cobra.Command{
	Use:   "edit <タイトル|ファイル名>",
	Short: "メモを編集",
	Long:  `指定したメモをエディタで開いて編集します。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		storage, err := note.NewStorage(viper.GetString("notes_dir"))
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
	Use:   "delete <タイトル|ファイル名>",
	Short: "メモを削除",
	Long:  `指定したメモを削除します。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		force, _ := cmd.Flags().GetBool("force")

		storage, err := note.NewStorage(viper.GetString("notes_dir"))
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

var noteSearchCmd = &cobra.Command{
	Use:   "search <クエリ>",
	Short: "メモを全文検索",
	Long:  `メモの内容を全文検索します。`,
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
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteShowCmd)
	noteCmd.AddCommand(noteEditCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteSearchCmd)

	noteCreateCmd.Flags().StringSliceP("tag", "t", []string{}, "タグを指定 (複数指定可)")
	noteCreateCmd.Flags().StringP("template", "T", "", "テンプレート名 (.templates/内のファイル)")
	noteListCmd.Flags().StringP("tag", "t", "", "タグでフィルタ")
	noteDeleteCmd.Flags().BoolP("force", "f", false, "確認なしで削除")
}

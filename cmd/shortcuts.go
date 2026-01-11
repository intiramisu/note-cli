package cmd

import (
	"fmt"
	"strings"

	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ルートレベルのショートカットコマンド（メモ操作をより短く）

var createCmd = &cobra.Command{
	Use:   "create <タイトル>",
	Short: "新規メモを作成 (note create のショートカット)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		tags, _ := cmd.Flags().GetStringSlice("tag")

		storage, err := note.NewStorage(viper.GetString("notes_dir"))
		if err != nil {
			return err
		}

		n := note.NewNote(title, tags)
		if err := storage.Save(n); err != nil {
			return err
		}

		fmt.Printf("メモを作成しました: %s\n", n.ID)
		return openEditor(storage.GetPath(n.ID))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "メモ一覧を表示 (note list のショートカット)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tagFilter, _ := cmd.Flags().GetString("tag")

		storage, err := note.NewStorage(viper.GetString("notes_dir"))
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
			fmt.Printf("- %s%s (%s)\n", n.Title, tagsStr, n.Modified.Format("2006-01-02 15:04"))
		}

		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <タイトル|ファイル名>",
	Short: "メモの内容を表示 (note show のショートカット)",
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

		fmt.Printf("# %s\n", n.Title)
		fmt.Printf("作成: %s | 更新: %s\n", n.Created.Format("2006-01-02 15:04"), n.Modified.Format("2006-01-02 15:04"))
		if len(n.Tags) > 0 {
			fmt.Printf("タグ: %s\n", strings.Join(n.Tags, ", "))
		}
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println(n.Content)

		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit <タイトル|ファイル名>",
	Short: "メモを編集 (note edit のショートカット)",
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

var searchCmd = &cobra.Command{
	Use:   "search <クエリ>",
	Short: "メモを全文検索 (note search のショートカット)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		results, err := search.Search(viper.GetString("notes_dir"), query)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Printf("「%s」に一致するメモはありません\n", query)
			return nil
		}

		fmt.Printf("「%s」の検索結果: %d件\n\n", query, len(results))

		currentFile := ""
		for _, r := range results {
			if r.Filename != currentFile {
				fmt.Printf("📄 %s\n", r.Title)
				currentFile = r.Filename
			}
			content := r.Content
			if len(content) > 80 {
				content = content[:77] + "..."
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

	createCmd.Flags().StringSliceP("tag", "t", []string{}, "タグを指定 (複数指定可)")
	listCmd.Flags().StringP("tag", "t", "", "タグでフィルタ")
}

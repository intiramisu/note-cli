package cmd

import "github.com/spf13/cobra"

// Root-level shortcut commands (shorter access to note operations)
// These reuse the RunE functions from note.go to avoid code duplication.

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new note",
	Args:  cobra.MinimumNArgs(1),
	RunE:  noteCreateCmd.RunE,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	RunE:  noteListCmd.RunE,
}

var showCmd = &cobra.Command{
	Use:   "show <title>",
	Short: "Show note content",
	Args:  cobra.MinimumNArgs(1),
	RunE:  noteShowCmd.RunE,
}

var editCmd = &cobra.Command{
	Use:   "edit <title>",
	Short: "Edit a note",
	Args:  cobra.MinimumNArgs(1),
	RunE:  noteEditCmd.RunE,
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)

	createCmd.Flags().StringSliceP("tag", "t", []string{}, "tags (can be specified multiple times)")
	createCmd.Flags().StringP("template", "T", "", "template name")
	listCmd.Flags().StringP("tag", "t", "", "filter by tag")
}

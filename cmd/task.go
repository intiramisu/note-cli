package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t"},
	Short:   "Manage tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}
		return task.Run(manager)
	},
}

var taskAddCmd = &cobra.Command{
	Use:   "add <description>",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := strings.Join(args, " ")
		priorityStr, _ := cmd.Flags().GetString("priority")
		noteID, _ := cmd.Flags().GetString("note")

		priority := task.PriorityNone
		switch priorityStr {
		case "1", "high":
			priority = task.PriorityHigh
		case "2", "medium":
			priority = task.PriorityMedium
		case "3", "low":
			priority = task.PriorityLow
		}

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		var t *task.Task
		if noteID != "" {
			t = manager.AddWithNote(description, priority, noteID)
			fmt.Printf("タスクを追加しました: [%d] %s (📄 %s)\n", t.ID, t.Description, noteID)
		} else {
			t = manager.Add(description, priority)
			fmt.Printf("タスクを追加しました: [%d] %s\n", t.ID, t.Description)
		}
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		showAll, _ := cmd.Flags().GetBool("all")

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		tasks := manager.List(showAll)
		if len(tasks) == 0 {
			fmt.Println("タスクがありません")
			return nil
		}

		for _, t := range tasks {
			checkbox := "[ ]"
			if t.IsDone() {
				checkbox = "[✓]"
			}
			priorityStr := ""
			if t.Priority != task.PriorityNone {
				priorityStr = fmt.Sprintf(" (%s)", t.Priority.String())
			}
			noteStr := ""
			if t.HasNote() {
				noteStr = fmt.Sprintf(" 📄 %s", t.NoteID)
			}
			fmt.Printf("%s [%d]%s %s%s\n", checkbox, t.ID, priorityStr, t.Description, noteStr)
		}

		return nil
	},
}

var taskDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無効なID: %s", args[0])
		}

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		if err := manager.Done(id); err != nil {
			return err
		}

		t, _ := manager.Get(id)
		fmt.Printf("タスクを完了しました: [%d] %s\n", t.ID, t.Description)
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("無効なID: %s", args[0])
		}

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		t, err := manager.Get(id)
		if err != nil {
			return err
		}
		desc := t.Description

		if err := manager.Delete(id); err != nil {
			return err
		}

		fmt.Printf("タスクを削除しました: [%d] %s\n", id, desc)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	taskAddCmd.Flags().StringP("priority", "p", "", "priority (1/high, 2/medium, 3/low)")
	taskAddCmd.Flags().StringP("note", "n", "", "link to a note")
	taskListCmd.Flags().BoolP("all", "a", false, "show completed tasks too")
}

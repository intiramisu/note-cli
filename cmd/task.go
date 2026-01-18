package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/spf13/cobra"
)

// parseDueDate parses flexible date formats
func parseDueDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	switch s {
	case "today":
		return today, nil
	case "tomorrow", "tom":
		return today.AddDate(0, 0, 1), nil
	}

	// +N days format
	if strings.HasPrefix(s, "+") {
		days, err := strconv.Atoi(s[1:])
		if err == nil {
			return today.AddDate(0, 0, days), nil
		}
	}

	// ISO format: 2026-01-20
	if t, err := time.ParseInLocation("2006-01-02", s, now.Location()); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location()), nil
	}

	// Short format: 01-20 (current year)
	if t, err := time.ParseInLocation("01-02", s, now.Location()); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", s)
}

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
		dueStr, _ := cmd.Flags().GetString("due")

		priority := task.PriorityNone
		switch priorityStr {
		case "1", "high":
			priority = task.PriorityHigh
		case "2", "medium":
			priority = task.PriorityMedium
		case "3", "low":
			priority = task.PriorityLow
		}

		dueDate, err := parseDueDate(dueStr)
		if err != nil {
			return err
		}

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		t := manager.AddFull(description, priority, noteID, dueDate)

		// 出力メッセージを構築
		var extras []string
		if noteID != "" {
			extras = append(extras, fmt.Sprintf("📄 %s", noteID))
		}
		if t.HasDueDate() {
			extras = append(extras, fmt.Sprintf("📅 %s", t.DueDate.Format("2006-01-02")))
		}
		extraStr := ""
		if len(extras) > 0 {
			extraStr = " (" + strings.Join(extras, ", ") + ")"
		}
		fmt.Printf("タスクを追加しました: [%d] %s%s\n", t.ID, t.Description, extraStr)
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		showAll, _ := cmd.Flags().GetBool("all")
		sortByDue, _ := cmd.Flags().GetBool("due")

		manager, err := task.NewManager(config.Global.NotesDir)
		if err != nil {
			return err
		}

		var tasks []*task.Task
		if sortByDue {
			tasks = manager.ListByDueDate(showAll)
		} else {
			tasks = manager.List(showAll)
		}

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
			dueStr := ""
			if t.HasDueDate() {
				dueLabel := t.DueDate.Format("01/02")
				if t.IsOverdue() {
					dueStr = fmt.Sprintf(" ⚠️ %s", dueLabel)
				} else if t.IsDueSoon(3) {
					dueStr = fmt.Sprintf(" 📅 %s", dueLabel)
				} else {
					dueStr = fmt.Sprintf(" 📅 %s", dueLabel)
				}
			}
			fmt.Printf("%s [%d]%s %s%s%s\n", checkbox, t.ID, priorityStr, t.Description, noteStr, dueStr)
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
	taskAddCmd.Flags().StringP("due", "d", "", "due date (2006-01-02, tomorrow, +3)")
	taskListCmd.Flags().BoolP("all", "a", false, "show completed tasks too")
	taskListCmd.Flags().BoolP("due", "d", false, "sort by due date")
}

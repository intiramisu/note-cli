package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/mattn/go-runewidth"
)

type viewMode int

const (
	modeNotesList viewMode = iota
	modeNoteDetail
)

type model struct {
	noteStorage *note.Storage
	taskManager *task.Manager

	mode         viewMode
	notes        []*note.Note
	selectedNote int
	tasks        []*task.Task
	selectedTask int

	width  int
	height int

	// タスク追加用
	addingTask bool
	taskInput  textinput.Model
	taskPriority task.Priority
}

func NewModel(noteStorage *note.Storage, taskManager *task.Manager) model {
	ti := textinput.New()
	ti.Placeholder = "タスクを入力..."
	ti.CharLimit = 200

	return model{
		noteStorage:  noteStorage,
		taskManager:  taskManager,
		mode:         modeNotesList,
		taskInput:    ti,
		taskPriority: task.PriorityMedium,
	}
}

func (m model) Init() tea.Cmd {
	return m.loadNotes
}

func (m *model) loadNotes() tea.Msg {
	notes, err := m.noteStorage.List("")
	if err != nil {
		return errMsg{err}
	}
	return notesLoadedMsg{notes}
}

type notesLoadedMsg struct {
	notes []*note.Note
}

type errMsg struct {
	err error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.addingTask {
			return m.handleTaskInput(msg)
		}
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case notesLoadedMsg:
		m.notes = msg.notes
		return m, nil

	case errMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		m.moveDown()

	case "k", "up":
		m.moveUp()

	case "enter":
		if m.mode == modeNotesList && len(m.notes) > 0 {
			m.mode = modeNoteDetail
			m.selectedTask = 0
			m.loadRelatedTasks()
		} else if m.mode == modeNoteDetail && len(m.tasks) > 0 {
			m.toggleTask()
		}

	case " ":
		if m.mode == modeNoteDetail && len(m.tasks) > 0 {
			m.toggleTask()
		}

	case "tab":
		if m.mode == modeNoteDetail {
			m.mode = modeNotesList
		}

	case "esc":
		if m.mode == modeNoteDetail {
			m.mode = modeNotesList
		}

	case "i":
		if m.mode == modeNoteDetail {
			m.addingTask = true
			m.taskInput.Focus()
			m.taskPriority = task.PriorityMedium
			return m, textinput.Blink
		}

	case "d", "x":
		if m.mode == modeNoteDetail && len(m.tasks) > 0 {
			m.deleteTask()
		}

	case "o":
		if m.mode == modeNoteDetail && len(m.tasks) > 0 {
			m.unlinkTask()
		}
	}

	return m, nil
}

func (m model) handleTaskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.taskInput.Value() != "" {
			m.addTask()
		}
		m.addingTask = false
		m.taskInput.Reset()
		return m, nil

	case "esc":
		m.addingTask = false
		m.taskInput.Reset()
		return m, nil

	case "tab":
		m.taskPriority = cyclePriority(m.taskPriority, false)
		return m, nil

	case "shift+tab":
		m.taskPriority = cyclePriority(m.taskPriority, true)
		return m, nil
	}

	var cmd tea.Cmd
	m.taskInput, cmd = m.taskInput.Update(msg)
	return m, cmd
}

func cyclePriority(p task.Priority, reverse bool) task.Priority {
	order := []task.Priority{task.PriorityHigh, task.PriorityMedium, task.PriorityLow}
	for i, pr := range order {
		if pr == p {
			if reverse {
				return order[(i-1+len(order))%len(order)]
			}
			return order[(i+1)%len(order)]
		}
	}
	return task.PriorityMedium
}

func (m *model) moveDown() {
	if m.mode == modeNotesList {
		if m.selectedNote < len(m.notes)-1 {
			m.selectedNote++
		}
	} else {
		if m.selectedTask < len(m.tasks)-1 {
			m.selectedTask++
		}
	}
}

func (m *model) moveUp() {
	if m.mode == modeNotesList {
		if m.selectedNote > 0 {
			m.selectedNote--
		}
	} else {
		if m.selectedTask > 0 {
			m.selectedTask--
		}
	}
}

func (m *model) loadRelatedTasks() {
	if m.selectedNote >= 0 && m.selectedNote < len(m.notes) {
		noteID := m.notes[m.selectedNote].ID
		m.tasks = m.taskManager.ListByNote(noteID)
	}
}

func (m *model) toggleTask() {
	if m.selectedTask >= 0 && m.selectedTask < len(m.tasks) {
		t := m.tasks[m.selectedTask]
		m.taskManager.Toggle(t.ID)
		m.loadRelatedTasks()
	}
}

func (m *model) deleteTask() {
	if m.selectedTask >= 0 && m.selectedTask < len(m.tasks) {
		t := m.tasks[m.selectedTask]
		m.taskManager.Delete(t.ID)
		m.loadRelatedTasks()
		if m.selectedTask >= len(m.tasks) && m.selectedTask > 0 {
			m.selectedTask--
		}
	}
}

func (m *model) unlinkTask() {
	if m.selectedTask >= 0 && m.selectedTask < len(m.tasks) {
		t := m.tasks[m.selectedTask]
		m.taskManager.UnlinkNote(t.ID)
		m.loadRelatedTasks()
		if m.selectedTask >= len(m.tasks) && m.selectedTask > 0 {
			m.selectedTask--
		}
	}
}

func (m *model) addTask() {
	if m.selectedNote >= 0 && m.selectedNote < len(m.notes) {
		noteID := m.notes[m.selectedNote].ID
		m.taskManager.AddWithNote(m.taskInput.Value(), m.taskPriority, noteID)
		m.loadRelatedTasks()
	}
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.mode {
	case modeNotesList:
		return m.renderNotesList()
	case modeNoteDetail:
		return m.renderNoteDetail()
	}
	return ""
}

func (m model) renderNotesList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var b strings.Builder
	b.WriteString(titleStyle.Render("📝 Notes"))
	b.WriteString("\n\n")

	if len(m.notes) == 0 {
		b.WriteString("メモがありません\n")
	} else {
		maxItems := m.height - 6
		if maxItems < 1 {
			maxItems = 1
		}

		start := 0
		if m.selectedNote >= maxItems {
			start = m.selectedNote - maxItems + 1
		}
		end := start + maxItems
		if end > len(m.notes) {
			end = len(m.notes)
		}

		for i := start; i < end; i++ {
			n := m.notes[i]
			prefix := "  "
			style := normalStyle
			if i == m.selectedNote {
				prefix = "▸ "
				style = selectedStyle
			}

			title := truncateString(n.Title, m.width-10)
			date := n.Modified.Format("2006-01-02")
			line := fmt.Sprintf("%s%-*s %s", prefix, m.width-15, title, date)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("j/k: 移動 | Enter: 詳細 | q: 終了"))

	return b.String()
}

func (m model) renderNoteDetail() string {
	if m.selectedNote < 0 || m.selectedNote >= len(m.notes) {
		return "メモが選択されていません"
	}

	n := m.notes[m.selectedNote]

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	taskTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginTop(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	doneStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Strikethrough(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var b strings.Builder

	// メモヘッダー
	b.WriteString(titleStyle.Render("📄 " + n.Title))
	b.WriteString("\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("作成: %s | 更新: %s",
		n.Created.Format("2006-01-02 15:04"),
		n.Modified.Format("2006-01-02 15:04"))))
	b.WriteString("\n")

	if len(n.Tags) > 0 {
		b.WriteString(metaStyle.Render("タグ: " + strings.Join(n.Tags, ", ")))
		b.WriteString("\n")
	}

	// メモ内容（最初の数行）
	b.WriteString(strings.Repeat("─", min(m.width-2, 40)))
	b.WriteString("\n")

	contentLines := strings.Split(n.Content, "\n")
	maxContentLines := (m.height - 15) / 2
	if maxContentLines < 3 {
		maxContentLines = 3
	}
	for i, line := range contentLines {
		if i >= maxContentLines {
			b.WriteString(metaStyle.Render("..."))
			b.WriteString("\n")
			break
		}
		b.WriteString(truncateString(line, m.width-4))
		b.WriteString("\n")
	}

	// 関連タスク
	b.WriteString("\n")
	b.WriteString(taskTitleStyle.Render("📋 関連タスク"))
	b.WriteString("\n")

	if m.addingTask {
		priorityLabel := m.taskPriority.String()
		b.WriteString(fmt.Sprintf("  [%s] %s\n", priorityLabel, m.taskInput.View()))
		b.WriteString(metaStyle.Render("  Tab: 優先度変更 | Enter: 確定 | Esc: キャンセル"))
		b.WriteString("\n")
	}

	if len(m.tasks) == 0 && !m.addingTask {
		b.WriteString(metaStyle.Render("  タスクなし"))
		b.WriteString("\n")
	} else {
		maxTaskLines := m.height - 15 - maxContentLines
		if maxTaskLines < 3 {
			maxTaskLines = 3
		}

		for i, t := range m.tasks {
			if i >= maxTaskLines {
				b.WriteString(metaStyle.Render(fmt.Sprintf("  ... 他 %d 件", len(m.tasks)-i)))
				b.WriteString("\n")
				break
			}

			prefix := "  "
			style := normalStyle
			if i == m.selectedTask && !m.addingTask {
				prefix = "▸ "
				style = selectedStyle
			}

			checkbox := "[ ]"
			if t.IsDone() {
				checkbox = "[✓]"
				style = doneStyle
			}

			priority := t.Priority.String()
			desc := truncateString(t.Description, m.width-20)
			line := fmt.Sprintf("%s%s (%s) %s", prefix, checkbox, priority, desc)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	if !m.addingTask {
		b.WriteString(helpStyle.Render("j/k: 移動 | Enter/Space: 完了切替 | i: 追加 | d: 削除 | o: 紐づけ解除 | Tab/Esc: 戻る"))
	}

	return b.String()
}

func truncateString(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	var result strings.Builder
	width := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if width+rw > maxWidth-3 {
			result.WriteString("...")
			break
		}
		result.WriteRune(r)
		width += rw
	}
	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Run(noteStorage *note.Storage, taskManager *task.Manager) error {
	p := tea.NewProgram(NewModel(noteStorage, taskManager), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/mattn/go-runewidth"
)

// UI styles (initialized from config)
type uiStyles struct {
	title    lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	done     lipgloss.Style
	meta     lipgloss.Style
	help     lipgloss.Style
}

var styles uiStyles

func initStyles() {
	cfg := config.Global
	colors := cfg.Theme.Colors

	styles = uiStyles{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Title)).MarginBottom(1),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Selected)).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#fcfcfc")),
		done:     lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)).Strikethrough(true),
		meta:     lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)).MarginTop(1),
	}
}

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
	addingTask   bool
	taskInput    textinput.Model
	taskPriority task.Priority
}

func NewModel(noteStorage *note.Storage, taskManager *task.Manager) model {
	initStyles()
	cfg := config.Global

	ti := textinput.New()
	ti.Placeholder = "タスクを入力..."
	ti.CharLimit = cfg.Display.TaskCharLimit
	ti.Width = cfg.Display.InputWidth
	ti.SetValue("")

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
			m.taskInput.Reset()
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
	cfg := config.Global
	symbols := cfg.Theme.Symbols
	formats := cfg.Formats

	var b strings.Builder
	b.WriteString(styles.title.Render(symbols.NoteIcon + " Notes"))
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
			prefix := symbols.CursorEmpty
			style := styles.normal
			if i == m.selectedNote {
				prefix = symbols.Cursor
				style = styles.selected
			}

			date := n.Modified.Format(formats.Date)
			dateWidth := runewidth.StringWidth(date)
			prefixWidth := runewidth.StringWidth(prefix)
			// タイトル用の幅 = 画面幅 - prefix幅 - 日付幅 - スペース2つ
			titleMaxWidth := m.width - prefixWidth - dateWidth - 2
			if titleMaxWidth < 10 {
				titleMaxWidth = 10
			}
			// サブディレクトリにあるノートはパスを表示
			dir := filepath.Dir(n.ID)
			titleDisplay := n.Title
			if dir != "." {
				titleDisplay = dir + "/" + n.Title
			}
			title := truncateString(titleDisplay, titleMaxWidth)
			// パディングを計算して右揃えの日付表示
			titleWidth := runewidth.StringWidth(title)
			padding := titleMaxWidth - titleWidth
			if padding < 0 {
				padding = 0
			}
			line := fmt.Sprintf("%s%s%s %s", prefix, title, strings.Repeat(" ", padding), date)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString(styles.help.Render("j/k: 移動 | Enter: 詳細 | q: 終了"))

	return b.String()
}

func (m model) renderNoteDetail() string {
	if m.selectedNote < 0 || m.selectedNote >= len(m.notes) {
		return "メモが選択されていません"
	}

	cfg := config.Global
	symbols := cfg.Theme.Symbols
	formats := cfg.Formats

	n := m.notes[m.selectedNote]

	var b strings.Builder

	// メモヘッダー
	b.WriteString(styles.title.Render(symbols.NoteIcon + " " + n.Title))
	b.WriteString("\n")
	b.WriteString(styles.meta.Render(fmt.Sprintf("作成: %s | 更新: %s",
		n.Created.Format(formats.DateTime),
		n.Modified.Format(formats.DateTime))))
	b.WriteString("\n")

	if len(n.Tags) > 0 {
		b.WriteString(styles.meta.Render("タグ: " + strings.Join(n.Tags, ", ")))
		b.WriteString("\n")
	}

	// メモ内容（最初の数行）
	sepWidth := cfg.Display.SeparatorWidth
	if sepWidth > m.width-2 {
		sepWidth = m.width - 2
	}
	b.WriteString(strings.Repeat("─", sepWidth))
	b.WriteString("\n")

	contentLines := strings.Split(n.Content, "\n")
	maxContentLines := (m.height - 15) / 2
	if maxContentLines < 3 {
		maxContentLines = 3
	}
	for i, line := range contentLines {
		if i >= maxContentLines {
			b.WriteString(styles.meta.Render("..."))
			b.WriteString("\n")
			break
		}
		b.WriteString(truncateString(line, m.width-4))
		b.WriteString("\n")
	}

	// 関連タスク
	b.WriteString("\n")
	taskTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Theme.Colors.Selected)).MarginTop(1)
	b.WriteString(taskTitleStyle.Render(symbols.TaskIcon + " 関連タスク"))
	b.WriteString("\n")

	if m.addingTask {
		priorityLabel := m.taskPriority.String()
		b.WriteString(fmt.Sprintf("  [%s] %s\n", priorityLabel, m.taskInput.View()))
		b.WriteString(styles.meta.Render("  Tab: 優先度変更 | Enter: 確定 | Esc: キャンセル"))
		b.WriteString("\n")
	}

	if len(m.tasks) == 0 && !m.addingTask {
		b.WriteString(styles.meta.Render("  タスクなし"))
		b.WriteString("\n")
	} else {
		maxTaskLines := m.height - 15 - maxContentLines
		if maxTaskLines < 3 {
			maxTaskLines = 3
		}

		for i, t := range m.tasks {
			if i >= maxTaskLines {
				b.WriteString(styles.meta.Render(fmt.Sprintf("  ... 他 %d 件", len(m.tasks)-i)))
				b.WriteString("\n")
				break
			}

			prefix := symbols.CursorEmpty
			style := styles.normal
			if i == m.selectedTask && !m.addingTask {
				prefix = symbols.Cursor
				style = styles.selected
			}

			checkbox := symbols.CheckboxEmpty
			if t.IsDone() {
				checkbox = symbols.CheckboxDone
				style = styles.done
			}

			priority := t.Priority.String()
			dueStr := ""
			if t.HasDueDate() {
				if t.IsOverdue() {
					dueStr = " ⚠️" + t.DueDate.Format("01/02")
				} else {
					dueStr = " 📅" + t.DueDate.Format("01/02")
				}
			}
			desc := truncateString(t.Description, m.width-25-len(dueStr))
			line := fmt.Sprintf("%s%s (%s) %s%s", prefix, checkbox, priority, desc, dueStr)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	if !m.addingTask {
		b.WriteString(styles.help.Render("j/k: 移動 | Enter/Space: 完了切替 | i: 追加 | d: 削除 | o: 紐づけ解除 | Tab/Esc: 戻る"))
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

func Run(noteStorage *note.Storage, taskManager *task.Manager) error {
	p := tea.NewProgram(NewModel(noteStorage, taskManager), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

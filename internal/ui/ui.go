package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/intiramisu/note-cli/internal/util"
	"github.com/mattn/go-runewidth"
)

var styles util.Styles

func initStyles() {
	styles = util.NewStyles(config.Global)
}

type viewMode int

const (
	modeNotesList viewMode = iota
	modeNoteDetail
	modeAttachTask
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

	// 期限入力用
	settingDue bool
	dueInput   textinput.Model
	taskDue    time.Time

	// ソート順
	sortByDue bool // true: 期限順, false: 優先度順

	// タスク紐づけ用
	unlinkedTasks    []*task.Task
	selectedUnlinked int
}

func NewModel(noteStorage *note.Storage, taskManager *task.Manager) model {
	initStyles()
	cfg := config.Global

	ti := textinput.New()
	ti.CharLimit = cfg.Display.TaskCharLimit
	ti.Width = cfg.Display.InputWidth
	ti.SetValue("")

	di := textinput.New()
	di.CharLimit = 20
	di.Width = 30
	di.SetValue("")

	return model{
		noteStorage:  noteStorage,
		taskManager:  taskManager,
		mode:         modeNotesList,
		taskInput:    ti,
		taskPriority: task.PriorityMedium,
		dueInput:     di,
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
		if m.mode == modeAttachTask {
			return m.handleAttachTask(msg)
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

	// Forward other messages (cursor blink etc.) to active text input
	if m.addingTask {
		var cmd tea.Cmd
		if m.settingDue {
			m.dueInput, cmd = m.dueInput.Update(msg)
		} else {
			m.taskInput, cmd = m.taskInput.Update(msg)
		}
		return m, cmd
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

	case "a":
		if m.mode == modeNoteDetail {
			m.loadUnlinkedTasks()
			if len(m.unlinkedTasks) > 0 {
				m.mode = modeAttachTask
				m.selectedUnlinked = 0
			}
		}

	case "s":
		if m.mode == modeNoteDetail {
			m.sortByDue = !m.sortByDue
			m.loadRelatedTasks()
		}
	}

	return m, nil
}

func (m model) handleTaskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 期限入力モード
	if m.settingDue {
		switch msg.String() {
		case "enter":
			if m.dueInput.Value() != "" {
				m.taskDue = util.ParseDueDateSimple(m.dueInput.Value())
			}
			m.settingDue = false
			m.addTask()
			m.addingTask = false
			m.taskInput.Reset()
			m.dueInput.Reset()
			m.taskDue = time.Time{}
			return m, nil

		case "esc":
			m.settingDue = false
			m.dueInput.Reset()
			m.taskInput.Focus()
			return m, textinput.Blink
		}

		var cmd tea.Cmd
		m.dueInput, cmd = m.dueInput.Update(msg)
		return m, cmd
	}

	// タスク説明入力モード
	switch msg.String() {
	case "enter":
		if m.taskInput.Value() != "" {
			m.addTask()
		}
		m.addingTask = false
		m.taskInput.Reset()
		m.taskDue = time.Time{}
		return m, nil

	case "esc":
		m.addingTask = false
		m.taskInput.Reset()
		m.taskDue = time.Time{}
		return m, nil

	case "tab":
		m.taskPriority = task.CyclePriority(m.taskPriority, false)
		return m, nil

	case "shift+tab":
		m.taskPriority = task.CyclePriority(m.taskPriority, true)
		return m, nil

	case "ctrl+d":
		if m.taskInput.Value() != "" {
			m.settingDue = true
			m.taskInput.Blur()
			m.dueInput.Focus()
			return m, textinput.Blink
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.taskInput, cmd = m.taskInput.Update(msg)
	return m, cmd
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

		if m.sortByDue {
			m.taskManager.SortByDueDate(m.tasks)
		}
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
		m.taskManager.Add(m.taskInput.Value(), m.taskPriority, noteID, m.taskDue)
		m.loadRelatedTasks()
	}
}


func (m *model) loadUnlinkedTasks() {
	allTasks := m.taskManager.List(false)
	m.unlinkedTasks = []*task.Task{}
	for _, t := range allTasks {
		if !t.HasNote() {
			m.unlinkedTasks = append(m.unlinkedTasks, t)
		}
	}
}

func (m model) handleAttachTask(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedUnlinked < len(m.unlinkedTasks)-1 {
			m.selectedUnlinked++
		}

	case "k", "up":
		if m.selectedUnlinked > 0 {
			m.selectedUnlinked--
		}

	case "enter":
		if m.selectedUnlinked >= 0 && m.selectedUnlinked < len(m.unlinkedTasks) {
			t := m.unlinkedTasks[m.selectedUnlinked]
			noteID := m.notes[m.selectedNote].ID
			m.taskManager.SetNoteID(t.ID, noteID)
			m.loadRelatedTasks()
			m.mode = modeNoteDetail
		}

	case "esc", "q":
		m.mode = modeNoteDetail
	}

	return m, nil
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
	case modeAttachTask:
		return m.renderAttachTask()
	}
	return ""
}

func (m model) renderNotesList() string {
	cfg := config.Global
	symbols := cfg.Theme.Symbols
	formats := cfg.Formats

	var b strings.Builder
	b.WriteString(styles.Title.Render(symbols.NoteIcon + " Notes"))
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
			style := styles.Normal
			if i == m.selectedNote {
				prefix = symbols.Cursor
				style = styles.Selected
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
			title := util.TruncateString(titleDisplay, titleMaxWidth)
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

	b.WriteString(styles.Help.Render("j/k: 移動 | Enter: 詳細 | q: 終了"))

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
	b.WriteString(styles.Title.Render(symbols.NoteIcon + " " + n.Title))
	b.WriteString("\n")
	b.WriteString(styles.Meta.Render(fmt.Sprintf("作成: %s | 更新: %s",
		n.Created.Format(formats.DateTime),
		n.Modified.Format(formats.DateTime))))
	b.WriteString("\n")

	if len(n.Tags) > 0 {
		b.WriteString(styles.Meta.Render("タグ: " + strings.Join(n.Tags, ", ")))
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
			b.WriteString(styles.Meta.Render("..."))
			b.WriteString("\n")
			break
		}
		b.WriteString(util.TruncateString(line, m.width-4))
		b.WriteString("\n")
	}

	// リンク情報
	links := note.ExtractLinks(n.Content)
	if len(links) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.Meta.Render("🔗 リンク先: "))
		found, notFound := note.ResolveLinks(m.noteStorage, links)
		var parts []string
		for _, ln := range found {
			parts = append(parts, ln.Title)
		}
		for _, name := range notFound {
			parts = append(parts, name+"(?)")
		}
		b.WriteString(styles.Meta.Render(strings.Join(parts, ", ")))
		b.WriteString("\n")
	}

	backlinks, _ := note.FindBacklinks(m.noteStorage, n.Title)
	if len(backlinks) > 0 {
		b.WriteString(styles.Meta.Render("🔙 被参照: "))
		var parts []string
		for _, bl := range backlinks {
			parts = append(parts, bl.Title)
		}
		b.WriteString(styles.Meta.Render(strings.Join(parts, ", ")))
		b.WriteString("\n")
	}

	// 関連タスク
	b.WriteString("\n")
	taskTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Theme.Colors.Selected)).MarginTop(1)
	b.WriteString(taskTitleStyle.Render(symbols.TaskIcon + " 関連タスク"))
	b.WriteString("\n")

	if m.addingTask {
		priorityLabel := m.taskPriority.String()
		if m.settingDue {
			// 期限入力モード
			b.WriteString(fmt.Sprintf("  [%s] %s\n", priorityLabel, m.taskInput.Value()))
			b.WriteString(fmt.Sprintf("  期限: %s\n", m.dueInput.View()))
			b.WriteString(styles.Meta.Render("  Enter: 確定 | Esc: 戻る"))
			b.WriteString("\n")
		} else {
			// タスク説明入力モード
			b.WriteString(fmt.Sprintf("  [%s] %s\n", priorityLabel, m.taskInput.View()))
			b.WriteString(styles.Meta.Render("  Tab: 優先度変更 | Ctrl+D: 期限設定 | Enter: 確定 | Esc: キャンセル"))
			b.WriteString("\n")
		}
	}

	if len(m.tasks) == 0 && !m.addingTask {
		b.WriteString(styles.Meta.Render("  タスクなし"))
		b.WriteString("\n")
	} else {
		maxTaskLines := m.height - 15 - maxContentLines
		if maxTaskLines < 3 {
			maxTaskLines = 3
		}

		for i, t := range m.tasks {
			if i >= maxTaskLines {
				b.WriteString(styles.Meta.Render(fmt.Sprintf("  ... 他 %d 件", len(m.tasks)-i)))
				b.WriteString("\n")
				break
			}

			prefix := symbols.CursorEmpty
			style := styles.Normal
			if i == m.selectedTask && !m.addingTask {
				prefix = symbols.Cursor
				style = styles.Selected
			}

			checkbox := symbols.CheckboxEmpty
			if t.IsDone() {
				checkbox = symbols.CheckboxDone
				style = styles.Done
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
			desc := util.TruncateString(t.Description, m.width-25-len(dueStr))
			line := fmt.Sprintf("%s%s (%s) %s%s", prefix, checkbox, priority, desc, dueStr)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	if !m.addingTask {
		sortLabel := "s: 期限順"
		if m.sortByDue {
			sortLabel = "s: 優先度順"
		}
		b.WriteString(styles.Help.Render(fmt.Sprintf("j/k: 移動 | Enter/Space: 完了切替 | i: 追加 | a: 紐づけ | d: 削除 | o: 解除 | %s | Tab/Esc: 戻る", sortLabel)))
	}

	return b.String()
}

func (m model) renderAttachTask() string {
	cfg := config.Global
	symbols := cfg.Theme.Symbols

	var b strings.Builder

	// タイトル
	n := m.notes[m.selectedNote]
	b.WriteString(styles.Title.Render(symbols.NoteIcon + " " + n.Title + " - タスクを紐づけ"))
	b.WriteString("\n\n")

	if len(m.unlinkedTasks) == 0 {
		b.WriteString(styles.Meta.Render("紐づけ可能なタスクがありません"))
		b.WriteString("\n")
	} else {
		b.WriteString(styles.Meta.Render("紐づけるタスクを選択:"))
		b.WriteString("\n\n")

		maxItems := m.height - 8
		if maxItems < 3 {
			maxItems = 3
		}

		start := 0
		if m.selectedUnlinked >= maxItems {
			start = m.selectedUnlinked - maxItems + 1
		}
		end := start + maxItems
		if end > len(m.unlinkedTasks) {
			end = len(m.unlinkedTasks)
		}

		for i := start; i < end; i++ {
			t := m.unlinkedTasks[i]
			prefix := symbols.CursorEmpty
			style := styles.Normal
			if i == m.selectedUnlinked {
				prefix = symbols.Cursor
				style = styles.Selected
			}

			priority := ""
			if t.Priority != task.PriorityNone {
				priority = fmt.Sprintf("(%s) ", t.Priority.String())
			}
			dueStr := ""
			if t.HasDueDate() {
				dueStr = fmt.Sprintf(" 📅%s", t.DueDate.Format("01/02"))
			}
			desc := util.TruncateString(t.Description, m.width-20)
			line := fmt.Sprintf("%s%s%s%s", prefix, priority, desc, dueStr)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.Help.Render("j/k: 移動 | Enter: 紐づけ | Esc: キャンセル"))

	return b.String()
}


func Run(noteStorage *note.Storage, taskManager *task.Manager) error {
	p := tea.NewProgram(NewModel(noteStorage, taskManager), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

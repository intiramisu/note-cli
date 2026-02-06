package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/util"
	"github.com/mattn/go-runewidth"
)

var styles util.Styles

func initStyles() {
	styles = util.NewStyles(config.Global)
}

type mode int

const (
	modeNormal mode = iota
	modeAdd
)

type sectionInfo struct {
	name     string
	color    string
	priority Priority
	isDone   bool
	tasks    []*Task
}

type Model struct {
	manager     *Manager
	sections    []sectionInfo
	sectionIdx  int
	taskIdx     int
	mode        mode
	textInput   textinput.Model
	addPriority Priority
	settingDue  bool
	dueInput    textinput.Model
	addDue      time.Time
	sortByDue   bool // true: 期限順, false: 優先度順
	quitting    bool
	width       int
	height      int
}

func NewModel(manager *Manager) Model {
	initStyles()
	cfg := config.Global

	ti := textinput.New()
	ti.Placeholder = "タスクの説明を入力 (Tabで優先度変更)..."
	ti.CharLimit = cfg.Display.TaskCharLimit
	ti.Width = cfg.Display.InputWidth

	di := textinput.New()
	di.Placeholder = "期限 (例: 01/20, 2026-01-20)"
	di.CharLimit = 20
	di.Width = 30

	m := Model{
		manager:     manager,
		textInput:   ti,
		dueInput:    di,
		addPriority: PriorityMedium,
		width:       120,
		height:      24,
	}
	m.refreshTasks()
	m.findFirstTask()
	return m
}

func (m *Model) findFirstTask() {
	for i, section := range m.sections {
		if len(section.tasks) > 0 {
			m.sectionIdx = i
			m.taskIdx = 0
			return
		}
	}
	m.sectionIdx = 0
	m.taskIdx = 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeAdd {
			return m.updateAddMode(msg)
		}
		return m.updateNormalMode(msg)
	}
	return m, nil
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		m.moveUp()

	case "down", "j":
		m.moveDown()

	case "left", "h":
		m.moveLeft()

	case "right", "l":
		m.moveRight()

	case "enter", " ":
		if task := m.currentTask(); task != nil {
			taskID := task.ID
			m.manager.Toggle(taskID)
			m.refreshTasks()
			m.moveCursorToTask(taskID)
		}

	case "i":
		m.mode = modeAdd
		m.addPriority = PriorityMedium
		m.textInput.Focus()
		return m, textinput.Blink

	case "d", "x":
		if task := m.currentTask(); task != nil {
			m.manager.Delete(task.ID)
			m.refreshTasks()
			m.adjustCursor()
		}

	case "s":
		m.sortByDue = !m.sortByDue
		m.refreshTasks()
		m.findFirstTask()
	}

	return m, nil
}

func (m *Model) moveUp() {
	if m.taskIdx > 0 {
		m.taskIdx--
	}
}

func (m *Model) moveDown() {
	if m.sectionIdx < len(m.sections) && m.taskIdx < len(m.sections[m.sectionIdx].tasks)-1 {
		m.taskIdx++
	}
}

func (m *Model) moveLeft() {
	for i := m.sectionIdx - 1; i >= 0; i-- {
		if len(m.sections[i].tasks) > 0 {
			m.sectionIdx = i
			if m.taskIdx >= len(m.sections[i].tasks) {
				m.taskIdx = len(m.sections[i].tasks) - 1
			}
			return
		}
	}
}

func (m *Model) moveRight() {
	for i := m.sectionIdx + 1; i < len(m.sections); i++ {
		if len(m.sections[i].tasks) > 0 {
			m.sectionIdx = i
			if m.taskIdx >= len(m.sections[i].tasks) {
				m.taskIdx = len(m.sections[i].tasks) - 1
			}
			return
		}
	}
}

func (m *Model) currentTask() *Task {
	if m.sectionIdx >= len(m.sections) {
		return nil
	}
	section := m.sections[m.sectionIdx]
	if m.taskIdx >= len(section.tasks) {
		return nil
	}
	return section.tasks[m.taskIdx]
}

func (m *Model) moveCursorToTask(taskID int) {
	for i, section := range m.sections {
		for j, t := range section.tasks {
			if t.ID == taskID {
				m.sectionIdx = i
				m.taskIdx = j
				return
			}
		}
	}
}

func (m *Model) adjustCursor() {
	if m.sectionIdx >= len(m.sections) {
		m.findFirstTask()
		return
	}
	section := m.sections[m.sectionIdx]
	if len(section.tasks) == 0 {
		m.moveRight()
		if m.currentTask() == nil {
			m.moveLeft()
		}
		if m.currentTask() == nil {
			m.findFirstTask()
		}
	} else if m.taskIdx >= len(section.tasks) {
		m.taskIdx = len(section.tasks) - 1
	}
}

func (m Model) updateAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 期限入力モード
	if m.settingDue {
		switch msg.String() {
		case "enter":
			if m.dueInput.Value() != "" {
				m.addDue = util.ParseDueDateSimple(m.dueInput.Value())
			}
			m.settingDue = false
			value := strings.TrimSpace(m.textInput.Value())
			if value != "" {
				newTask := m.manager.Add(value, m.addPriority, "", m.addDue)
				m.refreshTasks()
				m.moveCursorToTask(newTask.ID)
			}
			m.textInput.Reset()
			m.dueInput.Reset()
			m.mode = modeNormal
			m.addPriority = PriorityMedium
			m.addDue = time.Time{}
			return m, nil

		case "esc":
			m.settingDue = false
			m.dueInput.Reset()
			m.textInput.Focus()
			return m, textinput.Blink
		}

		var cmd tea.Cmd
		m.dueInput, cmd = m.dueInput.Update(msg)
		return m, cmd
	}

	// タスク説明入力モード
	switch msg.String() {
	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		if value != "" {
			newTask := m.manager.Add(value, m.addPriority, "", time.Time{})
			m.refreshTasks()
			m.moveCursorToTask(newTask.ID)
		}
		m.textInput.Reset()
		m.mode = modeNormal
		m.addPriority = PriorityMedium
		m.addDue = time.Time{}
		return m, nil

	case "esc":
		m.textInput.Reset()
		m.mode = modeNormal
		m.addPriority = PriorityMedium
		m.addDue = time.Time{}
		return m, nil

	case "tab":
		m.addPriority = CyclePriority(m.addPriority, false)
		return m, nil

	case "shift+tab":
		m.addPriority = CyclePriority(m.addPriority, true)
		return m, nil

	case "ctrl+d":
		if m.textInput.Value() != "" {
			m.settingDue = true
			m.textInput.Blur()
			m.dueInput.Focus()
			return m, textinput.Blink
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}


func (m *Model) refreshTasks() {
	cfg := config.Global
	colors := cfg.Theme.Colors
	sections := cfg.Theme.Sections

	if m.sortByDue {
		// 期限順表示: 1セクションに全タスク
		allTasks := m.manager.ListByDueDate(true)
		var pending, done []*Task
		for _, t := range allTasks {
			if t.IsDone() {
				done = append(done, t)
			} else {
				pending = append(pending, t)
			}
		}
		m.sections = []sectionInfo{
			{name: "📅 期限順", color: colors.Selected, tasks: pending},
			{name: sections.Done, color: colors.Done, isDone: true, tasks: done},
		}
	} else {
		// 優先度順表示: P1/P2/P3/Done
		allTasks := m.manager.List(true)
		m.sections = []sectionInfo{
			{name: sections.P1, color: colors.PriorityHigh, priority: PriorityHigh, tasks: []*Task{}},
			{name: sections.P2, color: colors.PriorityMedium, priority: PriorityMedium, tasks: []*Task{}},
			{name: sections.P3, color: colors.PriorityLow, priority: PriorityLow, tasks: []*Task{}},
			{name: sections.Done, color: colors.Done, isDone: true, tasks: []*Task{}},
		}

		for _, t := range allTasks {
			idx := m.sectionIndexForTask(t)
			m.sections[idx].tasks = append(m.sections[idx].tasks, t)
		}
	}
}

func (m *Model) sectionIndexForTask(t *Task) int {
	if t.IsDone() {
		return 3
	}
	switch t.Priority {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	default:
		return 2
	}
}

func (m Model) renderSection(sectionIndex int, section sectionInfo, colWidth, colHeight int) string {
	style := newSectionStyle(section.color, colWidth, colHeight)
	tStyle := m.sectionTitleStyle(section)

	var content strings.Builder
	content.WriteString(tStyle.Render(section.name))
	content.WriteString("\n")

	if len(section.tasks) == 0 {
		content.WriteString(styles.Empty.Render("  (なし)\n"))
		return style.Render(content.String())
	}

	for taskIndex, task := range section.tasks {
		isSelected := sectionIndex == m.sectionIdx && taskIndex == m.taskIdx
		line := m.renderTaskLine(task, colWidth, isSelected)
		content.WriteString(line + "\n")
	}

	return style.Render(content.String())
}

func (m Model) renderTaskLine(task *Task, colWidth int, isSelected bool) string {
	cfg := config.Global
	symbols := cfg.Theme.Symbols

	cursor := symbols.CursorEmpty
	if isSelected {
		cursor = symbols.Cursor
	}

	checkbox := symbols.CheckboxEmpty
	if task.IsDone() {
		checkbox = symbols.CheckboxDone
	}

	prefix := cursor + checkbox + " "
	prefixWidth := runewidth.StringWidth(prefix)
	maxDescWidth := max(colWidth-prefixWidth-4, 5)

	lines := util.WrapByWidth(task.Description, maxDescWidth)
	var result strings.Builder

	for i, line := range lines {
		if i == 0 {
			result.WriteString(prefix)
		} else {
			result.WriteString(strings.Repeat(" ", prefixWidth))
		}
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	// 期限がある場合は表示
	if task.HasDueDate() {
		result.WriteString("\n")
		dueLabel := "📅 " + task.DueDate.Format("01/02")
		if task.IsOverdue() {
			dueLabel = "⚠️ " + task.DueDate.Format("01/02")
		}
		result.WriteString(strings.Repeat(" ", prefixWidth))
		if task.IsOverdue() {
			result.WriteString(styles.PriorityHigh.Render(dueLabel))
		} else if task.IsDueSoon(3) {
			result.WriteString(styles.PriorityMedium.Render(dueLabel))
		} else {
			result.WriteString(styles.Help.Render(dueLabel))
		}
	}

	// 紐づきメモがある場合は表示
	if task.HasNote() {
		result.WriteString("\n")
		noteLabel := symbols.NoteIcon + " " + util.TruncateString(task.NoteID, maxDescWidth-3)
		result.WriteString(strings.Repeat(" ", prefixWidth))
		result.WriteString(styles.Help.Render(noteLabel))
	}

	text := result.String()
	if isSelected {
		return styles.Selected.Render(text)
	}
	if task.IsDone() {
		return styles.Done.Render(text)
	}
	return text
}


func (m Model) sectionTitleStyle(section sectionInfo) lipgloss.Style {
	if section.isDone {
		return styles.DoneSection
	}
	switch section.priority {
	case PriorityHigh:
		return styles.PriorityHigh
	case PriorityMedium:
		return styles.PriorityMedium
	case PriorityLow:
		return styles.PriorityLow
	}
	return styles.PriorityMedium
}

func (m Model) priorityStyle(p Priority) lipgloss.Style {
	switch p {
	case PriorityHigh:
		return styles.PriorityHigh
	case PriorityMedium:
		return styles.PriorityMedium
	case PriorityLow:
		return styles.PriorityLow
	}
	return styles.PriorityMedium
}

func newSectionStyle(color string, width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(0, 1).
		Width(width).
		Height(height)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	cfg := config.Global
	symbols := cfg.Theme.Symbols

	var s strings.Builder
	s.WriteString(styles.Title.Render(symbols.TaskIcon + " タスク管理"))
	s.WriteString("\n\n")

	colWidth, colHeight := m.calculateDimensions()
	s.WriteString(m.renderAllSections(colWidth, colHeight))
	s.WriteString("\n")

	if m.mode == modeAdd {
		s.WriteString(m.renderAddInput())
	}

	s.WriteString("\n")
	s.WriteString(styles.Help.Render(m.helpText()))
	return s.String()
}

func (m Model) calculateDimensions() (width, height int) {
	numSections := len(m.sections)
	width = max((m.width-numSections*2)/numSections, 15)
	height = max(m.height-8, 5)
	return
}

func (m Model) renderAllSections(colWidth, colHeight int) string {
	views := make([]string, len(m.sections))
	for i, section := range m.sections {
		views[i] = m.renderSection(i, section, colWidth, colHeight)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func (m Model) renderAddInput() string {
	style := m.priorityStyle(m.addPriority)
	label := style.Render("[" + m.addPriority.String() + "]")
	if m.settingDue {
		return fmt.Sprintf("\n新規タスク %s: %s\n期限: %s", label, m.textInput.Value(), m.dueInput.View())
	}
	return fmt.Sprintf("\n新規タスク %s: %s", label, m.textInput.View())
}

func (m Model) helpText() string {
	if m.mode == modeAdd {
		if m.settingDue {
			return "Enter:確定 Esc:戻る"
		}
		return "Enter:確定 Tab:優先度変更 Ctrl+D:期限設定 Esc:キャンセル"
	}
	sortLabel := "s:期限順"
	if m.sortByDue {
		sortLabel = "s:優先度順"
	}
	return fmt.Sprintf("i:追加 d:削除 Enter/Space:完了切替 %s h/l:左右 j/k:上下 q:終了", sortLabel)
}

func Run(manager *Manager) error {
	p := tea.NewProgram(NewModel(manager), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

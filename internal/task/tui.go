package task

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/config"
	"github.com/mattn/go-runewidth"
)

// TUI styles (initialized from config)
type tuiStyles struct {
	appTitle       lipgloss.Style
	selected       lipgloss.Style
	done           lipgloss.Style
	help           lipgloss.Style
	empty          lipgloss.Style
	priorityHigh   lipgloss.Style
	priorityMedium lipgloss.Style
	priorityLow    lipgloss.Style
	doneSection    lipgloss.Style
}

var styles tuiStyles

func initStyles() {
	cfg := config.Global
	colors := cfg.Theme.Colors

	styles = tuiStyles{
		appTitle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Title)).MarginBottom(1),
		selected:       lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Selected)).Bold(true),
		done:           lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)).Strikethrough(true),
		help:           lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)),
		empty:          lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Empty)),
		priorityHigh:   lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityHigh)).Bold(true),
		priorityMedium: lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityMedium)).Bold(true),
		priorityLow:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityLow)).Bold(true),
		doneSection:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)).Bold(true),
	}
}

var priorityCycle = []Priority{PriorityHigh, PriorityMedium, PriorityLow}

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

	m := Model{
		manager:     manager,
		textInput:   ti,
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
	switch msg.String() {
	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		if value != "" {
			newTask := m.manager.Add(value, m.addPriority)
			m.refreshTasks()
			m.moveCursorToTask(newTask.ID)
		}
		m.textInput.Reset()
		m.mode = modeNormal
		m.addPriority = PriorityMedium
		return m, nil

	case "esc":
		m.textInput.Reset()
		m.mode = modeNormal
		m.addPriority = PriorityMedium
		return m, nil

	case "tab":
		m.addPriority = cyclePriority(m.addPriority, false)
		return m, nil

	case "shift+tab":
		m.addPriority = cyclePriority(m.addPriority, true)
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func cyclePriority(current Priority, reverse bool) Priority {
	for i, p := range priorityCycle {
		if p == current {
			next := i + 1
			if reverse {
				next = i - 1 + len(priorityCycle)
			}
			return priorityCycle[next%len(priorityCycle)]
		}
	}
	return priorityCycle[0]
}

func (m *Model) refreshTasks() {
	allTasks := m.manager.List(true)
	cfg := config.Global
	colors := cfg.Theme.Colors
	sections := cfg.Theme.Sections

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
		content.WriteString(styles.empty.Render("  (なし)\n"))
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

	lines := wrapByWidth(task.Description, maxDescWidth)
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

	// 紐づきメモがある場合は表示
	if task.HasNote() {
		result.WriteString("\n")
		noteLabel := symbols.NoteIcon + " " + truncateByWidth(task.NoteID, maxDescWidth-3)
		result.WriteString(strings.Repeat(" ", prefixWidth))
		result.WriteString(styles.help.Render(noteLabel))
	}

	text := result.String()
	if isSelected {
		return styles.selected.Render(text)
	}
	if task.IsDone() {
		return styles.done.Render(text)
	}
	return text
}

func truncateByWidth(s string, maxWidth int) string {
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

func wrapByWidth(s string, maxWidth int) []string {
	if runewidth.StringWidth(s) <= maxWidth {
		return []string{s}
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}
		currentLine.WriteRune(r)
		currentWidth += rw
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

func (m Model) sectionTitleStyle(section sectionInfo) lipgloss.Style {
	if section.isDone {
		return styles.doneSection
	}
	switch section.priority {
	case PriorityHigh:
		return styles.priorityHigh
	case PriorityMedium:
		return styles.priorityMedium
	case PriorityLow:
		return styles.priorityLow
	}
	return styles.priorityMedium
}

func (m Model) priorityStyle(p Priority) lipgloss.Style {
	switch p {
	case PriorityHigh:
		return styles.priorityHigh
	case PriorityMedium:
		return styles.priorityMedium
	case PriorityLow:
		return styles.priorityLow
	}
	return styles.priorityMedium
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
	s.WriteString(styles.appTitle.Render(symbols.TaskIcon + " タスク管理"))
	s.WriteString("\n\n")

	colWidth, colHeight := m.calculateDimensions()
	s.WriteString(m.renderAllSections(colWidth, colHeight))
	s.WriteString("\n")

	if m.mode == modeAdd {
		s.WriteString(m.renderAddInput())
	}

	s.WriteString("\n")
	s.WriteString(styles.help.Render(m.helpText()))
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
	return fmt.Sprintf("\n新規タスク %s: %s", label, m.textInput.View())
}

func (m Model) helpText() string {
	if m.mode == modeAdd {
		return "Enter:確定 Tab:優先度変更 Esc:キャンセル"
	}
	return "i:追加 d:削除 Enter/Space:完了切替 h/l:左右 j/k:上下 q:終了"
}

func Run(manager *Manager) error {
	p := tea.NewProgram(NewModel(manager), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

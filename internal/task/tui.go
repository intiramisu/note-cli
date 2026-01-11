package task

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	appTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Strikethrough(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	emptyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	priorityStyles = map[Priority]lipgloss.Style{
		PriorityHigh:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		PriorityMedium: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		PriorityLow:    lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true),
	}

	priorityCycle = []Priority{PriorityHigh, PriorityMedium, PriorityLow}
)

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
	ti := textinput.New()
	ti.Placeholder = "タスクの説明を入力 (Tabで優先度変更)..."
	ti.CharLimit = 100
	ti.Width = 40

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

	m.sections = []sectionInfo{
		{name: "🔥 P1", color: "196", priority: PriorityHigh, tasks: []*Task{}},
		{name: "⚡ P2", color: "214", priority: PriorityMedium, tasks: []*Task{}},
		{name: "📝 P3", color: "75", priority: PriorityLow, tasks: []*Task{}},
		{name: "✅ 完了", color: "242", isDone: true, tasks: []*Task{}},
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
	tStyle := sectionTitleStyle(section)

	var content strings.Builder
	content.WriteString(tStyle.Render(section.name))
	content.WriteString("\n")

	if len(section.tasks) == 0 {
		content.WriteString(emptyStyle.Render("  (なし)\n"))
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
	cursor := "  "
	if isSelected {
		cursor = "▸ "
	}

	checkbox := "[ ]"
	if task.IsDone() {
		checkbox = "[✓]"
	}

	prefix := cursor + checkbox + " "
	prefixWidth := runewidth.StringWidth(prefix)
	maxDescWidth := max(colWidth-prefixWidth-4, 5) // padding分も考慮

	lines := wrapByWidth(task.Description, maxDescWidth)
	var result strings.Builder

	for i, line := range lines {
		if i == 0 {
			result.WriteString(prefix)
		} else {
			// 2行目以降はインデント
			result.WriteString(strings.Repeat(" ", prefixWidth))
		}
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	text := result.String()
	if isSelected {
		return selectedStyle.Render(text)
	}
	if task.IsDone() {
		return doneStyle.Render(text)
	}
	return text
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

func sectionTitleStyle(section sectionInfo) lipgloss.Style {
	if section.isDone {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Bold(true)
	}
	if style, ok := priorityStyles[section.priority]; ok {
		return style
	}
	return priorityStyles[PriorityMedium]
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

	var s strings.Builder
	s.WriteString(appTitleStyle.Render("📋 タスク管理"))
	s.WriteString("\n\n")

	colWidth, colHeight := m.calculateDimensions()
	s.WriteString(m.renderAllSections(colWidth, colHeight))
	s.WriteString("\n")

	if m.mode == modeAdd {
		s.WriteString(m.renderAddInput())
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render(m.helpText()))
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
	style := priorityStyles[m.addPriority]
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

package util

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/intiramisu/note-cli/internal/config"
)

// Styles holds common TUI styles.
type Styles struct {
	Title          lipgloss.Style
	Selected       lipgloss.Style
	Normal         lipgloss.Style
	Done           lipgloss.Style
	Meta           lipgloss.Style
	Help           lipgloss.Style
	Empty          lipgloss.Style
	PriorityHigh   lipgloss.Style
	PriorityMedium lipgloss.Style
	PriorityLow    lipgloss.Style
	DoneSection    lipgloss.Style
}

// NewStyles creates TUI styles from config.
func NewStyles(cfg *config.Config) Styles {
	colors := cfg.Theme.Colors

	return Styles{
		Title:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Title)).MarginBottom(1),
		Selected:       lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Selected)).Bold(true),
		Normal:         lipgloss.NewStyle().Foreground(lipgloss.Color("#fcfcfc")),
		Done:           lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)).Strikethrough(true),
		Meta:           lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)),
		Help:           lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Help)).MarginTop(1),
		Empty:          lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Empty)),
		PriorityHigh:   lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityHigh)).Bold(true),
		PriorityMedium: lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityMedium)).Bold(true),
		PriorityLow:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.PriorityLow)).Bold(true),
		DoneSection:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)).Bold(true),
	}
}

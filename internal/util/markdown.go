package util

import (
	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown content with ANSI styling for terminal output.
// Falls back to raw content on error.
func RenderMarkdown(content string, width int, style string) (string, error) {
	if width < 20 {
		width = 20
	}

	if style == "" {
		style = "dark"
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content, err
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return content, err
	}

	return rendered, nil
}

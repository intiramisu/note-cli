package util

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// TruncateString truncates a string to fit within maxWidth,
// adding "..." suffix if truncated. Properly handles wide characters.
func TruncateString(s string, maxWidth int) string {
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

// WrapByWidth wraps a string into multiple lines based on display width.
// Properly handles wide characters (CJK).
func WrapByWidth(s string, maxWidth int) []string {
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

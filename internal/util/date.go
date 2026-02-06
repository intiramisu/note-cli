package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDueDate parses flexible date formats for task due dates.
// Supports:
//   - "today", "tomorrow", "tom"
//   - "+N" (N days from today)
//   - "2006-01-02" (ISO format)
//   - "01-02", "01/02", "1/2" (current year)
func ParseDueDate(s string) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return time.Time{}, nil
	}

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

	// ISO format: 2006-01-02
	if t, err := time.ParseInLocation("2006-01-02", s, now.Location()); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location()), nil
	}

	// Short format with dash: 01-02 (current year)
	if t, err := time.ParseInLocation("01-02", s, now.Location()); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location()), nil
	}

	// Short format with slash: 01/02 (current year)
	if t, err := time.Parse("01/02", s); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
	}

	// Short format with slash: 1/2 (current year)
	if t, err := time.Parse("1/2", s); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", s)
}

// ParseDueDateSimple parses date without returning an error (returns zero time on failure).
// Used in TUI where we don't need detailed error messages.
func ParseDueDateSimple(s string) time.Time {
	t, _ := ParseDueDate(s)
	return t
}

// ParseDate parses flexible date inputs for daily notes.
// Supports: "today", "yesterday", "tomorrow", "+N", "-N", ISO date format.
func ParseDate(input string, dateFormat string) (time.Time, error) {
	now := time.Now()

	switch strings.ToLower(input) {
	case "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	}

	// +N / -N 形式
	if len(input) > 0 && (input[0] == '+' || input[0] == '-') {
		var days int
		if _, err := fmt.Sscanf(input, "%d", &days); err == nil {
			return now.AddDate(0, 0, days), nil
		}
	}

	// 日付形式 (設定から取得)
	parsed, err := time.Parse(dateFormat, input)
	if err != nil {
		return time.Time{}, fmt.Errorf("無効な日付形式: %s (%s, yesterday, tomorrow, +N, -N が使えます)", input, dateFormat)
	}
	return parsed, nil
}

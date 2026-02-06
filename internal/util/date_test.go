package util

import (
	"testing"
	"time"
)

func TestParseDueDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			name:  "empty string returns zero time",
			input: "",
			check: func(t *testing.T, got time.Time) {
				if !got.IsZero() {
					t.Errorf("expected zero time, got %v", got)
				}
			},
		},
		{
			name:  "today",
			input: "today",
			check: func(t *testing.T, got time.Time) {
				if !got.Equal(today) {
					t.Errorf("expected %v, got %v", today, got)
				}
			},
		},
		{
			name:  "tomorrow",
			input: "tomorrow",
			check: func(t *testing.T, got time.Time) {
				expected := today.AddDate(0, 0, 1)
				if !got.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, got)
				}
			},
		},
		{
			name:  "tom (short for tomorrow)",
			input: "tom",
			check: func(t *testing.T, got time.Time) {
				expected := today.AddDate(0, 0, 1)
				if !got.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, got)
				}
			},
		},
		{
			name:  "+3 days",
			input: "+3",
			check: func(t *testing.T, got time.Time) {
				expected := today.AddDate(0, 0, 3)
				if !got.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, got)
				}
			},
		},
		{
			name:  "ISO format 2026-03-15",
			input: "2026-03-15",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2026 || got.Month() != 3 || got.Day() != 15 {
					t.Errorf("expected 2026-03-15, got %v", got)
				}
			},
		},
		{
			name:  "short format 01-15",
			input: "01-15",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != now.Year() || got.Month() != 1 || got.Day() != 15 {
					t.Errorf("expected %d-01-15, got %v", now.Year(), got)
				}
			},
		},
		{
			name:  "short format 01/15",
			input: "01/15",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != now.Year() || got.Month() != 1 || got.Day() != 15 {
					t.Errorf("expected %d-01-15, got %v", now.Year(), got)
				}
			},
		},
		{
			name:  "short format 1/5",
			input: "1/5",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != now.Year() || got.Month() != 1 || got.Day() != 5 {
					t.Errorf("expected %d-01-05, got %v", now.Year(), got)
				}
			},
		},
		{
			name:    "invalid input",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDueDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseDueDateSimple(t *testing.T) {
	// Valid input
	got := ParseDueDateSimple("today")
	if got.IsZero() {
		t.Error("ParseDueDateSimple(today) returned zero time")
	}

	// Invalid input returns zero time
	got = ParseDueDateSimple("invalid")
	if !got.IsZero() {
		t.Errorf("ParseDueDateSimple(invalid) = %v, want zero time", got)
	}
}

func TestParseDate(t *testing.T) {
	now := time.Now()
	dateFormat := "2006-01-02"

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			name:  "today",
			input: "today",
			check: func(t *testing.T, got time.Time) {
				if got.Day() != now.Day() || got.Month() != now.Month() {
					t.Errorf("expected today, got %v", got)
				}
			},
		},
		{
			name:  "yesterday",
			input: "yesterday",
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -1)
				if got.Day() != expected.Day() || got.Month() != expected.Month() {
					t.Errorf("expected yesterday, got %v", got)
				}
			},
		},
		{
			name:  "tomorrow",
			input: "tomorrow",
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, 1)
				if got.Day() != expected.Day() || got.Month() != expected.Month() {
					t.Errorf("expected tomorrow, got %v", got)
				}
			},
		},
		{
			name:  "+3 days",
			input: "+3",
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, 3)
				if got.Day() != expected.Day() {
					t.Errorf("expected +3 days, got %v", got)
				}
			},
		},
		{
			name:  "-2 days",
			input: "-2",
			check: func(t *testing.T, got time.Time) {
				expected := now.AddDate(0, 0, -2)
				if got.Day() != expected.Day() {
					t.Errorf("expected -2 days, got %v", got)
				}
			},
		},
		{
			name:  "ISO date",
			input: "2026-01-15",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2026 || got.Month() != 1 || got.Day() != 15 {
					t.Errorf("expected 2026-01-15, got %v", got)
				}
			},
		},
		{
			name:    "invalid input",
			input:   "foobar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input, dateFormat)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

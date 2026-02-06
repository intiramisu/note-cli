package util

import (
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{
			name:     "ASCII short string (no truncation)",
			input:    "hello",
			maxWidth: 10,
			want:     "hello",
		},
		{
			name:     "ASCII exact width",
			input:    "hello",
			maxWidth: 5,
			want:     "hello",
		},
		{
			name:     "ASCII truncated",
			input:    "hello world",
			maxWidth: 8,
			want:     "hello...",
		},
		{
			name:     "Japanese no truncation",
			input:    "テスト",
			maxWidth: 10,
			want:     "テスト",
		},
		{
			name:     "Japanese exact width",
			input:    "テスト",
			maxWidth: 6,
			want:     "テスト",
		},
		{
			name:     "Japanese truncated",
			input:    "テストメモです",
			maxWidth: 10,
			want:     "テスト...",
		},
		{
			name:     "very small maxWidth",
			input:    "hello world",
			maxWidth: 4,
			want:     "h...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestWrapByWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     int // expected number of lines
	}{
		{
			name:     "short string (no wrapping)",
			input:    "hello",
			maxWidth: 10,
			want:     1,
		},
		{
			name:     "exact width (no wrapping)",
			input:    "hello",
			maxWidth: 5,
			want:     1,
		},
		{
			name:     "ASCII wrapping",
			input:    "hello world!",
			maxWidth: 5,
			want:     3,
		},
		{
			name:     "Japanese wrapping",
			input:    "テストメモです",
			maxWidth: 6,
			want:     3,
		},
		{
			name:     "fits within width",
			input:    "hi",
			maxWidth: 10,
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapByWidth(tt.input, tt.maxWidth)
			if len(got) != tt.want {
				t.Errorf("WrapByWidth(%q, %d) returned %d lines, want %d (lines: %v)",
					tt.input, tt.maxWidth, len(got), tt.want, got)
			}
		})
	}
}

package note

import (
	"time"
)

type Note struct {
	ID       string    `yaml:"-"`
	Title    string    `yaml:"title"`
	Created  time.Time `yaml:"created"`
	Modified time.Time `yaml:"modified"`
	Tags     []string  `yaml:"tags"`
	Content  string    `yaml:"-"`
}

func NewNote(title string, tags []string) *Note {
	now := time.Now()
	return &Note{
		Title:    title,
		Created:  now,
		Modified: now,
		Tags:     tags,
		Content:  "",
	}
}

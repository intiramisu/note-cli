package cmd

import (
	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/task"
)

func newStorage() (*note.Storage, error) {
	return note.NewStorage(config.Global.NotesDir)
}

func newTaskManager() (*task.Manager, error) {
	return task.NewManager(config.Global.NotesDir)
}

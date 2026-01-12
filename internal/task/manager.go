package task

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/intiramisu/note-cli/internal/config"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	filePath string
	tasks    []*Task
	nextID   int
}

func NewManager(notesDir string) (*Manager, error) {
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("ディレクトリの作成に失敗: %w", err)
	}

	tasksFile := ".tasks.yaml"
	if config.Global != nil && config.Global.Paths.TasksFile != "" {
		tasksFile = config.Global.Paths.TasksFile
	}

	m := &Manager{
		filePath: filepath.Join(notesDir, tasksFile),
		tasks:    []*Task{},
		nextID:   1,
	}

	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

func (m *Manager) Add(description string, priority Priority) *Task {
	task := NewTask(m.nextID, description, priority)
	m.tasks = append(m.tasks, task)
	m.nextID++
	m.save()
	return task
}

func (m *Manager) AddWithNote(description string, priority Priority, noteID string) *Task {
	task := NewTask(m.nextID, description, priority)
	task.NoteID = noteID
	m.tasks = append(m.tasks, task)
	m.nextID++
	m.save()
	return task
}

func (m *Manager) List(showDone bool) []*Task {
	var result []*Task
	for _, t := range m.tasks {
		if showDone || !t.IsDone() {
			result = append(result, t)
		}
	}
	return m.sortByPriority(result)
}

func (m *Manager) ListByNote(noteID string) []*Task {
	var result []*Task
	for _, t := range m.tasks {
		if t.NoteID == noteID {
			result = append(result, t)
		}
	}
	return m.sortByPriority(result)
}

func (m *Manager) sortByPriority(tasks []*Task) []*Task {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].Created.Before(tasks[j].Created)
	})
	return tasks
}

func (m *Manager) Get(id int) (*Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("タスクが見つかりません: ID=%d", id)
}

func (m *Manager) Done(id int) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}
	task.Done()
	return m.save()
}

func (m *Manager) Delete(id int) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return m.save()
		}
	}
	return fmt.Errorf("タスクが見つかりません: ID=%d", id)
}

func (m *Manager) Toggle(id int) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}
	if task.IsDone() {
		task.Status = StatusPending
		task.Completed = time.Time{}
	} else {
		task.Done()
	}
	return m.save()
}

func (m *Manager) SetNoteID(id int, noteID string) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}
	task.NoteID = noteID
	return m.save()
}

func (m *Manager) UnlinkNote(id int) error {
	return m.SetNoteID(id, "")
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var stored struct {
		NextID int     `yaml:"next_id"`
		Tasks  []*Task `yaml:"tasks"`
	}

	if err := yaml.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("タスクファイルのパースに失敗: %w", err)
	}

	m.tasks = stored.Tasks
	m.nextID = stored.NextID
	return nil
}

func (m *Manager) save() error {
	stored := struct {
		NextID int     `yaml:"next_id"`
		Tasks  []*Task `yaml:"tasks"`
	}{
		NextID: m.nextID,
		Tasks:  m.tasks,
	}

	data, err := yaml.Marshal(stored)
	if err != nil {
		return fmt.Errorf("タスクのシリアライズに失敗: %w", err)
	}

	return os.WriteFile(m.filePath, data, 0644)
}

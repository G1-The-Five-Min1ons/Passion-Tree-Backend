package handler

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskProcessing TaskStatus = "processing"
	TaskDone       TaskStatus = "done"
	TaskFailed     TaskStatus = "failed"
)

type Task struct {
	TaskID    string      `json:"task_id"`
	Status    TaskStatus  `json:"status"`
	StartedAt time.Time   `json:"started_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func newTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]*Task)}
}

func (ts *TaskStore) create() *Task {
	t := &Task{
		TaskID:    uuid.New().String(),
		Status:    TaskProcessing,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	ts.mu.Lock()
	ts.tasks[t.TaskID] = t
	ts.mu.Unlock()
	return t
}

func (ts *TaskStore) get(id string) (*Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tasks[id]
	if !ok {
		return nil, false
	}
	copyTask := *t
	return &copyTask, true
}

func (ts *TaskStore) complete(id string, result interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tasks[id]; ok {
		t.Status = TaskDone
		t.Result = result
		t.UpdatedAt = time.Now()
	}
}

func (ts *TaskStore) fail(id string, errMsg string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tasks[id]; ok {
		t.Status = TaskFailed
		t.Error = errMsg
		t.UpdatedAt = time.Now()
	}
}

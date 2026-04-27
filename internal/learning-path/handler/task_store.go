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
	TaskID    string     `json:"task_id"`
	Type      string     `json:"type"`
	Status    TaskStatus `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Result    interface{} `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type TaskStore struct {
	mu          sync.RWMutex
	tasks       map[string]*Task
	activeTypes map[string]bool 
}

func newTaskStore() *TaskStore {
	ts := &TaskStore{
		tasks:       make(map[string]*Task),
		activeTypes: make(map[string]bool),
	}

	go ts.startCleanupWorker(1 * time.Hour) 
	return ts
}


func (ts *TaskStore) create(taskType string) (*Task, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// ถ้าประเภทงานนี้กำลังรันอยู่ (เช่น BulkSync รันค้างไว้) ให้ปฏิเสธการสร้าง
	if ts.activeTypes[taskType] {
		return nil, false
	}

	t := &Task{
		TaskID:    uuid.New().String(),
		Type:      taskType,
		Status:    TaskProcessing,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	ts.tasks[t.TaskID] = t
	ts.activeTypes[taskType] = true
	return t, true
}

func (ts *TaskStore) get(id string) (*Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tasks[id]
	return t, ok
}

func (ts *TaskStore) complete(id string, result interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if t, ok := ts.tasks[id]; ok {
		t.Status = TaskDone
		t.Result = result
		t.UpdatedAt = time.Now()
		delete(ts.activeTypes, t.Type)
	}
}

func (ts *TaskStore) fail(id string, errMsg string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tasks[id]; ok {
		t.Status = TaskFailed
		t.Error = errMsg
		t.UpdatedAt = time.Now()
		delete(ts.activeTypes, t.Type) // Remove the task type from active types
	}
}

func (ts *TaskStore) startCleanupWorker(retention time.Duration) {
	ticker := time.NewTicker(30 * time.Minute) // ตรวจทุก 30 นาที
	for range ticker.C {
		ts.mu.Lock()
		for id, t := range ts.tasks {
			if t.Status != TaskProcessing && time.Since(t.UpdatedAt) > retention {
                delete(ts.tasks, id)
			}
		}
		ts.mu.Unlock()
	}
}
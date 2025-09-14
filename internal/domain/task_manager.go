package domain

import (
	"fmt"
	"sync"
	"time"
)

// TaskManager 任务管理器
type TaskManager struct {
	tasks map[string]*Task
	mutex sync.RWMutex
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

// CreateTask 创建任务
func (tm *TaskManager) CreateTask(prompt string) *Task {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task := &Task{
		ID:        generateTaskID(),
		Status:    TaskStatusPending,
		Prompt:    prompt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tm.tasks[task.ID] = task
	return task
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	task, exists := tm.tasks[taskID]
	return task, exists
}

// UpdateTaskStatus 更新任务状态
func (tm *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = status
		task.UpdatedAt = time.Now()
	}
}

// UpdateTaskResult 更新任务结果
func (tm *TaskManager) UpdateTaskResult(taskID string, result *ImageGenerationResponse) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = TaskStatusCompleted
		task.Result = result
		task.UpdatedAt = time.Now()
	}
}

// UpdateTaskError 更新任务错误
func (tm *TaskManager) UpdateTaskError(taskID string, errorMsg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = TaskStatusFailed
		task.Error = errorMsg
		task.UpdatedAt = time.Now()
	}
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

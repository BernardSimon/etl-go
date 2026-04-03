package task

import (
	"context"
	"errors"
	"sync"
)

// RunState provides thread-safe access to task execution state,
// protecting the cancel function map and manual cancel map from concurrent access.
type RunState struct {
	mu        sync.RWMutex
	ctxMap    map[string]context.CancelFunc
	cancelMap map[string]string
}

func NewRunState() *RunState {
	return &RunState{
		ctxMap:    make(map[string]context.CancelFunc),
		cancelMap: make(map[string]string),
	}
}

// SetCancel registers a cancel function for a task record.
func (rs *RunState) SetCancel(recordID string, cancel context.CancelFunc) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.ctxMap[recordID] = cancel
}

// RemoveCancel removes the cancel function for a task record.
func (rs *RunState) RemoveCancel(recordID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.ctxMap, recordID)
}

// CancelRecord cancels a running task record by its ID.
// Returns an error if the record is not found or not cancellable.
func (rs *RunState) CancelRecord(recordID string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	cancel, ok := rs.ctxMap[recordID]
	if !ok {
		return errors.New("任务不存在或状态不可停止")
	}
	cancel()
	rs.cancelMap[recordID] = "cancel"
	return nil
}

// IsManualCancelled checks if a task record was manually cancelled and clears the flag.
func (rs *RunState) IsManualCancelled(recordID string) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.cancelMap[recordID]; exists {
		delete(rs.cancelMap, recordID)
		return true
	}
	return false
}

// global singleton
var runState = NewRunState()

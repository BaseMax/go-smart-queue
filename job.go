package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// JobStatus represents the current state of a job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusRunning
	JobStatusCompleted
	JobStatusFailed
	JobStatusCancelled
)

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusRunning:
		return "running"
	case JobStatusCompleted:
		return "completed"
	case JobStatusFailed:
		return "failed"
	case JobStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// ResourceConstraints defines the resource limits for a job
type ResourceConstraints struct {
	MaxCPUPercent    float64 // Maximum CPU usage percentage (0-100)
	MaxMemoryMB      float64 // Maximum memory usage in MB
	MaxIOOpsPerSec   int     // Maximum IO operations per second
}

// Job represents a task to be executed
type Job struct {
	ID          string
	Name        string
	Priority    int // Higher value = higher priority
	Constraints ResourceConstraints
	Task        func(context.Context) error
	Status      JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       error
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewJob creates a new job with the given parameters
func NewJob(name string, priority int, constraints ResourceConstraints, task func(context.Context) error) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		ID:          generateID(),
		Name:        name,
		Priority:    priority,
		Constraints: constraints,
		Task:        task,
		Status:      JobStatusPending,
		CreatedAt:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Cancel cancels the job
func (j *Job) Cancel() {
	j.cancel()
	j.Status = JobStatusCancelled
}

var jobIDCounter uint64

func generateID() string {
	id := atomic.AddUint64(&jobIDCounter, 1)
	return fmt.Sprintf("job-%d", id)
}

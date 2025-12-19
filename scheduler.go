package main

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// JobQueue is a priority queue of jobs
type JobQueue []*Job

func (jq JobQueue) Len() int { return len(jq) }

func (jq JobQueue) Less(i, j int) bool {
	// Higher priority comes first
	return jq[i].Priority > jq[j].Priority
}

func (jq JobQueue) Swap(i, j int) {
	jq[i], jq[j] = jq[j], jq[i]
}

func (jq *JobQueue) Push(x interface{}) {
	*jq = append(*jq, x.(*Job))
}

func (jq *JobQueue) Pop() interface{} {
	old := *jq
	n := len(old)
	job := old[n-1]
	*jq = old[0 : n-1]
	return job
}

// Scheduler manages job execution with resource monitoring
type Scheduler struct {
	mu               sync.RWMutex
	queue            *JobQueue
	activeJobs       map[string]*Job
	completedJobs    []*Job
	maxWorkers       int
	activeWorkers    int
	monitor          *ResourceMonitor
	jobCh            chan *Job
	stopCh           chan struct{}
	stopped          bool
	checkInterval    time.Duration
}

// NewScheduler creates a new job scheduler
func NewScheduler(maxWorkers int, monitor *ResourceMonitor) *Scheduler {
	jq := &JobQueue{}
	heap.Init(jq)
	
	return &Scheduler{
		queue:         jq,
		activeJobs:    make(map[string]*Job),
		completedJobs: make([]*Job, 0),
		maxWorkers:    maxWorkers,
		monitor:       monitor,
		jobCh:         make(chan *Job, maxWorkers),
		stopCh:        make(chan struct{}),
		checkInterval: 500 * time.Millisecond,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	go s.dispatchLoop()
	for i := 0; i < s.maxWorkers; i++ {
		go s.workerLoop()
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.stopped {
		close(s.stopCh)
		s.stopped = true
	}
	s.mu.Unlock()
}

// AddJob adds a job to the queue
func (s *Scheduler) AddJob(job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return fmt.Errorf("scheduler is stopped")
	}

	heap.Push(s.queue, job)
	return nil
}

// CancelJob cancels a job by ID
func (s *Scheduler) CancelJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job is active
	if job, exists := s.activeJobs[jobID]; exists {
		job.Cancel()
		return nil
	}

	// Check if job is in queue
	for i := 0; i < s.queue.Len(); i++ {
		if (*s.queue)[i].ID == jobID {
			job := (*s.queue)[i]
			job.Cancel()
			heap.Remove(s.queue, i)
			s.completedJobs = append(s.completedJobs, job)
			return nil
		}
	}

	return fmt.Errorf("job not found: %s", jobID)
}

// GetJob returns a job by ID
func (s *Scheduler) GetJob(jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check active jobs
	if job, exists := s.activeJobs[jobID]; exists {
		return job, nil
	}

	// Check queue
	for i := 0; i < s.queue.Len(); i++ {
		if (*s.queue)[i].ID == jobID {
			return (*s.queue)[i], nil
		}
	}

	// Check completed jobs
	for _, job := range s.completedJobs {
		if job.ID == jobID {
			return job, nil
		}
	}

	return nil, fmt.Errorf("job not found: %s", jobID)
}

// ListJobs returns all jobs
func (s *Scheduler) ListJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0)

	// Add queued jobs
	for i := 0; i < s.queue.Len(); i++ {
		jobs = append(jobs, (*s.queue)[i])
	}

	// Add active jobs
	for _, job := range s.activeJobs {
		jobs = append(jobs, job)
	}

	// Add completed jobs (last 50)
	start := 0
	if len(s.completedJobs) > 50 {
		start = len(s.completedJobs) - 50
	}
	jobs = append(jobs, s.completedJobs[start:]...)

	return jobs
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"queued_jobs":    s.queue.Len(),
		"active_jobs":    len(s.activeJobs),
		"completed_jobs": len(s.completedJobs),
		"max_workers":    s.maxWorkers,
		"active_workers": s.activeWorkers,
	}
}

func (s *Scheduler) dispatchLoop() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tryDispatchJob()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) tryDispatchJob() {
	s.mu.Lock()
	
	// Check if we have capacity
	if s.activeWorkers >= s.maxWorkers || s.queue.Len() == 0 {
		s.mu.Unlock()
		return
	}

	// Get the highest priority job
	job := heap.Pop(s.queue).(*Job)
	
	// Check if system resources allow running this job
	if !s.monitor.CanRunJob(job.Constraints) {
		// Put the job back in the queue
		heap.Push(s.queue, job)
		s.mu.Unlock()
		return
	}

	// Dispatch the job
	s.activeJobs[job.ID] = job
	s.activeWorkers++
	s.mu.Unlock()

	// Send job to worker
	select {
	case s.jobCh <- job:
	case <-s.stopCh:
		return
	}
}

func (s *Scheduler) workerLoop() {
	for {
		select {
		case job := <-s.jobCh:
			s.executeJob(job)
		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) executeJob(job *Job) {
	defer func() {
		s.mu.Lock()
		delete(s.activeJobs, job.ID)
		s.completedJobs = append(s.completedJobs, job)
		s.activeWorkers--
		s.mu.Unlock()
	}()

	// Update job status
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now

	// Execute the job
	err := job.Task(job.ctx)

	// Update job status
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	if err != nil {
		if job.ctx.Err() == context.Canceled {
			job.Status = JobStatusCancelled
		} else {
			job.Status = JobStatusFailed
			job.Error = err
		}
	} else {
		job.Status = JobStatusCompleted
	}
}

package main

import (
	"context"
	"testing"
	"time"
)

func TestJobCreation(t *testing.T) {
	task := func(ctx context.Context) error {
		return nil
	}

	constraints := ResourceConstraints{
		MaxCPUPercent:  80.0,
		MaxMemoryMB:    1024.0,
		MaxIOOpsPerSec: 1000,
	}

	job := NewJob("test-job", 5, constraints, task)

	if job.Name != "test-job" {
		t.Errorf("Expected job name 'test-job', got '%s'", job.Name)
	}

	if job.Priority != 5 {
		t.Errorf("Expected priority 5, got %d", job.Priority)
	}

	if job.Status != JobStatusPending {
		t.Errorf("Expected status pending, got %v", job.Status)
	}

	if job.ID == "" {
		t.Error("Expected job ID to be generated")
	}
}

func TestJobCancel(t *testing.T) {
	task := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	constraints := ResourceConstraints{}
	job := NewJob("cancel-test", 5, constraints, task)

	job.Cancel()

	if job.Status != JobStatusCancelled {
		t.Errorf("Expected status cancelled, got %v", job.Status)
	}
}

func TestSchedulerAddJob(t *testing.T) {
	monitor := NewResourceMonitor(1 * time.Second)
	scheduler := NewScheduler(2, monitor)

	task := func(ctx context.Context) error {
		return nil
	}

	constraints := ResourceConstraints{}
	job := NewJob("test-job", 5, constraints, task)

	err := scheduler.AddJob(job)
	if err != nil {
		t.Errorf("Failed to add job: %v", err)
	}

	stats := scheduler.GetStats()
	if stats["queued_jobs"].(int) != 1 {
		t.Errorf("Expected 1 queued job, got %d", stats["queued_jobs"].(int))
	}
}

func TestSchedulerJobExecution(t *testing.T) {
	monitor := NewResourceMonitor(1 * time.Second)
	monitor.Start()
	defer monitor.Stop()

	scheduler := NewScheduler(2, monitor)
	scheduler.Start()
	defer scheduler.Stop()

	executed := false
	task := func(ctx context.Context) error {
		executed = true
		return nil
	}

	constraints := ResourceConstraints{
		MaxCPUPercent:  100.0,
		MaxMemoryMB:    10000.0,
		MaxIOOpsPerSec: 100000,
	}
	job := NewJob("test-job", 5, constraints, task)

	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Wait for job to execute
	time.Sleep(2 * time.Second)

	if !executed {
		t.Error("Job was not executed")
	}

	retrievedJob, err := scheduler.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Status != JobStatusCompleted {
		t.Errorf("Expected status completed, got %v", retrievedJob.Status)
	}
}

func TestSchedulerCancelJob(t *testing.T) {
	monitor := NewResourceMonitor(1 * time.Second)
	scheduler := NewScheduler(1, monitor)

	task := func(ctx context.Context) error {
		time.Sleep(10 * time.Second)
		return nil
	}

	constraints := ResourceConstraints{}
	job := NewJob("cancel-test", 5, constraints, task)

	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	err = scheduler.CancelJob(job.ID)
	if err != nil {
		t.Errorf("Failed to cancel job: %v", err)
	}

	retrievedJob, err := scheduler.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Status != JobStatusCancelled {
		t.Errorf("Expected status cancelled, got %v", retrievedJob.Status)
	}
}

func TestPriorityQueue(t *testing.T) {
	monitor := NewResourceMonitor(1 * time.Second)
	scheduler := NewScheduler(1, monitor)

	task := func(ctx context.Context) error {
		return nil
	}

	constraints := ResourceConstraints{}

	// Add jobs with different priorities
	job1 := NewJob("low-priority", 1, constraints, task)
	job2 := NewJob("high-priority", 10, constraints, task)
	job3 := NewJob("medium-priority", 5, constraints, task)

	scheduler.AddJob(job1)
	scheduler.AddJob(job2)
	scheduler.AddJob(job3)

	stats := scheduler.GetStats()
	if stats["queued_jobs"].(int) != 3 {
		t.Errorf("Expected 3 queued jobs, got %d", stats["queued_jobs"].(int))
	}

	// Verify jobs exist
	if _, err := scheduler.GetJob(job1.ID); err != nil {
		t.Errorf("Failed to find job1: %v", err)
	}
	if _, err := scheduler.GetJob(job2.ID); err != nil {
		t.Errorf("Failed to find job2: %v", err)
	}
	if _, err := scheduler.GetJob(job3.ID); err != nil {
		t.Errorf("Failed to find job3: %v", err)
	}
}

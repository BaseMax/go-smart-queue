package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Example demonstrates using the job scheduler programmatically
func ExampleUsage() {
	// Initialize resource monitor
	monitor := NewResourceMonitor(1 * time.Second)
	monitor.Start()
	defer monitor.Stop()

	// Initialize scheduler with 4 workers
	scheduler := NewScheduler(4, monitor)
	scheduler.Start()
	defer scheduler.Stop()

	// Give monitor time to collect initial metrics
	time.Sleep(2 * time.Second)

	fmt.Println("Starting job scheduler example...")
	fmt.Println()

	// Create different types of jobs
	jobs := []*Job{
		NewJob("Data Processing", 20, ResourceConstraints{
			MaxCPUPercent:  90.0,
			MaxMemoryMB:    7000.0,
			MaxIOOpsPerSec: 10000,
		}, createSampleTask("Processing data", 2*time.Second)),

		NewJob("Email Notifications", 15, ResourceConstraints{
			MaxCPUPercent:  90.0,
			MaxMemoryMB:    7000.0,
			MaxIOOpsPerSec: 10000,
		}, createSampleTask("Sending emails", 1*time.Second)),

		NewJob("Database Backup", 10, ResourceConstraints{
			MaxCPUPercent:  90.0,
			MaxMemoryMB:    7000.0,
			MaxIOOpsPerSec: 10000,
		}, createSampleTask("Backing up database", 3*time.Second)),

		NewJob("Cache Cleanup", 5, ResourceConstraints{
			MaxCPUPercent:  90.0,
			MaxMemoryMB:    7000.0,
			MaxIOOpsPerSec: 10000,
		}, createSampleTask("Cleaning cache", 1*time.Second)),

		NewJob("Report Generation", 18, ResourceConstraints{
			MaxCPUPercent:  90.0,
			MaxMemoryMB:    7000.0,
			MaxIOOpsPerSec: 10000,
		}, createSampleTask("Generating reports", 2*time.Second)),
	}

	// Add all jobs to the scheduler
	for _, job := range jobs {
		if err := scheduler.AddJob(job); err != nil {
			fmt.Printf("Error adding job %s: %v\n", job.Name, err)
		} else {
			fmt.Printf("Added job: %s (ID: %s, Priority: %d)\n", job.Name, job.ID, job.Priority)
		}
	}

	fmt.Println()
	fmt.Println("Jobs added to queue. Monitoring execution...")
	fmt.Println()

	// Monitor job execution
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(15 * time.Second)
	
	for {
		select {
		case <-ticker.C:
			stats := scheduler.GetStats()
			metrics := monitor.GetMetrics()
			
			fmt.Printf("[%s] Stats: Queued=%d, Active=%d, Completed=%d | CPU=%.1f%%, Mem=%.0fMB\n",
				time.Now().Format("15:04:05"),
				stats["queued_jobs"],
				stats["active_jobs"],
				stats["completed_jobs"],
				metrics.CPUPercent,
				metrics.MemoryUsedMB,
			)

			// Check if all jobs are done
			if stats["completed_jobs"].(int) >= len(jobs) {
				fmt.Println()
				fmt.Println("All jobs completed!")
				printFinalResults(scheduler, jobs)
				return
			}

		case <-timeout:
			fmt.Println()
			fmt.Println("Timeout reached. Stopping...")
			printFinalResults(scheduler, jobs)
			return
		}
	}
}

func createSampleTask(description string, duration time.Duration) func(context.Context) error {
	return func(ctx context.Context) error {
		fmt.Printf("  → Started: %s\n", description)
		
		// Simulate work with possible cancellation
		select {
		case <-time.After(duration):
			// Simulate occasional failures
			if rand.Float32() < 0.1 { // 10% failure rate
				return fmt.Errorf("simulated error in %s", description)
			}
			fmt.Printf("  ✓ Completed: %s\n", description)
			return nil
		case <-ctx.Done():
			fmt.Printf("  ✗ Cancelled: %s\n", description)
			return ctx.Err()
		}
	}
}

func printFinalResults(scheduler *Scheduler, jobs []*Job) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("Final Results")
	fmt.Println("========================================")
	fmt.Println()

	for _, job := range jobs {
		retrievedJob, err := scheduler.GetJob(job.ID)
		if err != nil {
			fmt.Printf("Error getting job %s: %v\n", job.ID, err)
			continue
		}

		duration := "N/A"
		if retrievedJob.StartedAt != nil && retrievedJob.CompletedAt != nil {
			duration = retrievedJob.CompletedAt.Sub(*retrievedJob.StartedAt).Round(time.Millisecond).String()
		}

		status := "✓"
		if retrievedJob.Status == JobStatusFailed {
			status = "✗"
		} else if retrievedJob.Status == JobStatusCancelled {
			status = "◯"
		}

		fmt.Printf("%s %-25s Priority: %2d  Status: %-10s  Duration: %s\n",
			status,
			retrievedJob.Name,
			retrievedJob.Priority,
			retrievedJob.Status.String(),
			duration,
		)

		if retrievedJob.Error != nil {
			fmt.Printf("  Error: %v\n", retrievedJob.Error)
		}
	}

	stats := scheduler.GetStats()
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total Jobs: %d\n", len(jobs))
	fmt.Printf("  Completed: %d\n", stats["completed_jobs"])
	fmt.Printf("  Max Workers Used: %d\n", stats["max_workers"])
	fmt.Println()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ExampleUsage()
}

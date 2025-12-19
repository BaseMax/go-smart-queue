package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	scheduler *Scheduler
	monitor   *ResourceMonitor
)

func main() {
	// Initialize monitor and scheduler
	monitor = NewResourceMonitor(1 * time.Second)
	monitor.Start()
	
	scheduler = NewScheduler(4, monitor)
	scheduler.Start()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		scheduler.Stop()
		monitor.Stop()
		os.Exit(0)
	}()

	rootCmd := &cobra.Command{
		Use:   "go-smart-queue",
		Short: "A smart job scheduler with resource monitoring",
		Long:  `A Go job scheduler that runs tasks while monitoring CPU, memory, and IO usage.`,
	}

	rootCmd.AddCommand(
		addJobCmd(),
		listJobsCmd(),
		statusCmd(),
		cancelJobCmd(),
		dashboardCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func addJobCmd() *cobra.Command {
	var (
		name           string
		priority       int
		maxCPU         float64
		maxMemory      float64
		maxIO          int
		command        string
		sleepDuration  int
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new job to the queue",
		Long:  `Add a new job with specified constraints and priority.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("job name is required")
			}

			constraints := ResourceConstraints{
				MaxCPUPercent:  maxCPU,
				MaxMemoryMB:    maxMemory,
				MaxIOOpsPerSec: maxIO,
			}

			// Create a sample task
			task := createTask(command, sleepDuration)

			job := NewJob(name, priority, constraints, task)
			if err := scheduler.AddJob(job); err != nil {
				return err
			}

			fmt.Printf("Job added successfully: %s (ID: %s)\n", job.Name, job.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Job name (required)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 5, "Job priority (higher = higher priority)")
	cmd.Flags().Float64Var(&maxCPU, "max-cpu", 80.0, "Maximum CPU percentage")
	cmd.Flags().Float64Var(&maxMemory, "max-memory", 1024.0, "Maximum memory in MB")
	cmd.Flags().IntVar(&maxIO, "max-io", 1000, "Maximum IO ops per second")
	cmd.Flags().StringVarP(&command, "command", "c", "", "Command to execute (optional)")
	cmd.Flags().IntVarP(&sleepDuration, "sleep", "s", 5, "Sleep duration in seconds (for demo)")

	cmd.MarkFlagRequired("name")

	return cmd
}

func listJobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all jobs",
		Long:  `List all jobs in the queue, running, and completed.`,
		Run: func(cmd *cobra.Command, args []string) {
			jobs := scheduler.ListJobs()
			
			if len(jobs) == 0 {
				fmt.Println("No jobs found.")
				return
			}

			fmt.Printf("%-15s %-20s %-10s %-12s %-15s\n", "ID", "NAME", "PRIORITY", "STATUS", "DURATION")
			fmt.Println(strings.Repeat("-", 80))

			for _, job := range jobs {
				duration := "N/A"
				if job.StartedAt != nil {
					endTime := time.Now()
					if job.CompletedAt != nil {
						endTime = *job.CompletedAt
					}
					duration = endTime.Sub(*job.StartedAt).Round(time.Second).String()
				}

				fmt.Printf("%-15s %-20s %-10d %-12s %-15s\n",
					job.ID,
					truncateString(job.Name, 20),
					job.Priority,
					job.Status.String(),
					duration,
				)
			}
		},
	}
}

func statusCmd() *cobra.Command {
	var jobID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of a job or overall system",
		Long:  `Show detailed status of a specific job or overall system statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jobID != "" {
				return showJobStatus(jobID)
			}
			return showSystemStatus()
		},
	}

	cmd.Flags().StringVarP(&jobID, "job", "j", "", "Job ID to show status for")

	return cmd
}

func cancelJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel [job-id]",
		Short: "Cancel a job",
		Long:  `Cancel a pending or running job by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]
			if err := scheduler.CancelJob(jobID); err != nil {
				return err
			}
			fmt.Printf("Job %s cancelled successfully.\n", jobID)
			return nil
		},
	}
}

func dashboardCmd() *cobra.Command {
	var refreshInterval int

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show real-time dashboard",
		Long:  `Display a real-time dashboard with system metrics and job status.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDashboard(time.Duration(refreshInterval) * time.Second)
		},
	}

	cmd.Flags().IntVarP(&refreshInterval, "refresh", "r", 2, "Refresh interval in seconds")

	return cmd
}

func createTask(command string, sleepDuration int) func(context.Context) error {
	return func(ctx context.Context) error {
		if command != "" {
			fmt.Printf("Executing command: %s\n", command)
		}
		
		// Simulate work
		select {
		case <-time.After(time.Duration(sleepDuration) * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func showJobStatus(jobID string) error {
	job, err := scheduler.GetJob(jobID)
	if err != nil {
		return err
	}

	fmt.Printf("Job Details:\n")
	fmt.Printf("  ID:          %s\n", job.ID)
	fmt.Printf("  Name:        %s\n", job.Name)
	fmt.Printf("  Priority:    %d\n", job.Priority)
	fmt.Printf("  Status:      %s\n", job.Status.String())
	fmt.Printf("  Created:     %s\n", job.CreatedAt.Format(time.RFC3339))
	
	if job.StartedAt != nil {
		fmt.Printf("  Started:     %s\n", job.StartedAt.Format(time.RFC3339))
	}
	
	if job.CompletedAt != nil {
		fmt.Printf("  Completed:   %s\n", job.CompletedAt.Format(time.RFC3339))
		duration := job.CompletedAt.Sub(*job.StartedAt)
		fmt.Printf("  Duration:    %s\n", duration.Round(time.Second))
	}

	fmt.Printf("\nResource Constraints:\n")
	fmt.Printf("  Max CPU:     %.2f%%\n", job.Constraints.MaxCPUPercent)
	fmt.Printf("  Max Memory:  %.2f MB\n", job.Constraints.MaxMemoryMB)
	fmt.Printf("  Max IO:      %d ops/sec\n", job.Constraints.MaxIOOpsPerSec)

	if job.Error != nil {
		fmt.Printf("\nError: %v\n", job.Error)
	}

	return nil
}

func showSystemStatus() error {
	stats := scheduler.GetStats()
	metrics := monitor.GetMetrics()

	fmt.Println("System Status:")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Println("\nScheduler Statistics:")
	fmt.Printf("  Queued Jobs:     %d\n", stats["queued_jobs"])
	fmt.Printf("  Active Jobs:     %d\n", stats["active_jobs"])
	fmt.Printf("  Completed Jobs:  %d\n", stats["completed_jobs"])
	fmt.Printf("  Max Workers:     %d\n", stats["max_workers"])
	fmt.Printf("  Active Workers:  %d\n", stats["active_workers"])

	fmt.Println("\nSystem Metrics:")
	fmt.Printf("  CPU Usage:       %.2f%%\n", metrics.CPUPercent)
	fmt.Printf("  Memory Used:     %.2f MB (%.2f%%)\n", metrics.MemoryUsedMB, metrics.MemoryPercent)
	fmt.Printf("  IO Read:         %d bytes\n", metrics.IOReadBytes)
	fmt.Printf("  IO Write:        %d bytes\n", metrics.IOWriteBytes)
	fmt.Printf("  Goroutines:      %d\n", GetGoRoutineCount())
	fmt.Printf("  Last Updated:    %s\n", metrics.Timestamp.Format(time.RFC3339))

	return nil
}

func runDashboard(refreshInterval time.Duration) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// Clear screen function
	clearScreen := func() {
		fmt.Print("\033[H\033[2J")
	}

	for {
		clearScreen()
		
		fmt.Println("╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║          Go Smart Queue - Real-Time Dashboard                 ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		fmt.Println()

		showSystemStatus()

		fmt.Println()
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println("Recent Jobs:")
		fmt.Println(strings.Repeat("─", 60))

		jobs := scheduler.ListJobs()
		displayCount := 10
		if len(jobs) < displayCount {
			displayCount = len(jobs)
		}

		if displayCount > 0 {
			fmt.Printf("%-12s %-20s %-8s %-12s\n", "ID", "NAME", "PRIORITY", "STATUS")
			fmt.Println(strings.Repeat("-", 60))

			for i := 0; i < displayCount; i++ {
				job := jobs[len(jobs)-displayCount+i]
				fmt.Printf("%-12s %-20s %-8d %-12s\n",
					job.ID,
					truncateString(job.Name, 20),
					job.Priority,
					job.Status.String(),
				)
			}
		} else {
			fmt.Println("No jobs to display.")
		}

		fmt.Println()
		fmt.Printf("Press Ctrl+C to exit | Refresh: %v\n", refreshInterval)

		select {
		case <-ticker.C:
			continue
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

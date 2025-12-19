package main

import (
"context"
"fmt"
"time"
)

func main() {
// Initialize resource monitor
monitor := NewResourceMonitor(500 * time.Millisecond)
monitor.Start()
defer monitor.Stop()

// Wait for metrics
time.Sleep(2 * time.Second)

// Check metrics
metrics := monitor.GetMetrics()
fmt.Printf("Metrics: CPU=%.2f%%, Mem=%.2fMB, Timestamp=%v\n", 
metrics.CPUPercent, metrics.MemoryUsedMB, metrics.Timestamp)

// Test constraints
constraints := ResourceConstraints{
MaxCPUPercent:  80.0,
MaxMemoryMB:    1024.0,
MaxIOOpsPerSec: 1000,
}

canRun := monitor.CanRunJob(constraints)
fmt.Printf("Can run job: %v\n", canRun)

// Initialize scheduler
scheduler := NewScheduler(4, monitor)
scheduler.Start()
defer scheduler.Stop()

// Add a simple job
job := NewJob("Test Job", 10, constraints, func(ctx context.Context) error {
fmt.Println("Job is running!")
time.Sleep(2 * time.Second)
fmt.Println("Job completed!")
return nil
})

err := scheduler.AddJob(job)
if err != nil {
fmt.Printf("Error adding job: %v\n", err)
}

fmt.Printf("Job added: %s\n", job.ID)

// Wait and monitor
for i := 0; i < 10; i++ {
time.Sleep(1 * time.Second)
stats := scheduler.GetStats()
fmt.Printf("Stats: Queued=%d, Active=%d, Completed=%d\n",
stats["queued_jobs"], stats["active_jobs"], stats["completed_jobs"])
}
}

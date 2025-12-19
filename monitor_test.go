package main

import (
	"testing"
	"time"
)

func TestResourceMonitorCreation(t *testing.T) {
	monitor := NewResourceMonitor(1 * time.Second)
	if monitor == nil {
		t.Fatal("Failed to create resource monitor")
	}

	if monitor.updateInterval != 1*time.Second {
		t.Errorf("Expected update interval 1s, got %v", monitor.updateInterval)
	}
}

func TestResourceMonitorStartStop(t *testing.T) {
	monitor := NewResourceMonitor(100 * time.Millisecond)
	monitor.Start()

	// Let it run for a bit
	time.Sleep(500 * time.Millisecond)

	// Get metrics
	metrics := monitor.GetMetrics()
	if metrics.Timestamp.IsZero() {
		t.Error("Expected metrics to be updated")
	}

	monitor.Stop()

	// Verify it can be called multiple times
	monitor.Stop()
}

func TestResourceMonitorMetrics(t *testing.T) {
	monitor := NewResourceMonitor(100 * time.Millisecond)
	monitor.Start()
	defer monitor.Stop()

	// Wait for metrics to be collected
	time.Sleep(500 * time.Millisecond)

	metrics := monitor.GetMetrics()

	// CPU should be a valid percentage
	if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
		t.Errorf("Invalid CPU percentage: %f", metrics.CPUPercent)
	}

	// Memory should be positive
	if metrics.MemoryUsedMB < 0 {
		t.Errorf("Invalid memory usage: %f", metrics.MemoryUsedMB)
	}

	// Memory percentage should be valid
	if metrics.MemoryPercent < 0 || metrics.MemoryPercent > 100 {
		t.Errorf("Invalid memory percentage: %f", metrics.MemoryPercent)
	}
}

func TestCanRunJobWithConstraints(t *testing.T) {
	monitor := NewResourceMonitor(100 * time.Millisecond)
	monitor.Start()
	defer monitor.Stop()

	// Wait for initial metrics
	time.Sleep(500 * time.Millisecond)

	// Test with very high constraints (should always pass)
	highConstraints := ResourceConstraints{
		MaxCPUPercent:  100.0,
		MaxMemoryMB:    100000.0,
		MaxIOOpsPerSec: 1000000,
	}

	if !monitor.CanRunJob(highConstraints) {
		t.Error("Should be able to run job with high constraints")
	}

	// Test with very low constraints (likely to fail)
	lowConstraints := ResourceConstraints{
		MaxCPUPercent:  0.01,
		MaxMemoryMB:    0.01,
		MaxIOOpsPerSec: 1,
	}

	canRun := monitor.CanRunJob(lowConstraints)
	// This might pass or fail depending on system state, just verify it returns a boolean
	_ = canRun
}

func TestGetGoRoutineCount(t *testing.T) {
	count := GetGoRoutineCount()
	if count <= 0 {
		t.Errorf("Expected positive goroutine count, got %d", count)
	}

	// Start some goroutines
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()
	}

	newCount := GetGoRoutineCount()
	if newCount <= count {
		t.Logf("Warning: Expected more goroutines after spawning. Before: %d, After: %d", count, newCount)
	}

	// Wait for goroutines to finish
	for i := 0; i < 5; i++ {
		<-done
	}
}

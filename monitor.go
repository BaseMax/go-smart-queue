package main

import (
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemMetrics holds the current system resource usage
type SystemMetrics struct {
	CPUPercent    float64
	MemoryUsedMB  float64
	MemoryPercent float64
	IOReadBytes   uint64
	IOWriteBytes  uint64
	IOReadOps     uint64
	IOWriteOps    uint64
	Timestamp     time.Time
}

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	mu              sync.RWMutex
	currentMetrics  SystemMetrics
	previousMetrics SystemMetrics
	updateInterval  time.Duration
	stopCh          chan struct{}
	stopped         bool
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(updateInterval time.Duration) *ResourceMonitor {
	return &ResourceMonitor{
		updateInterval: updateInterval,
		stopCh:         make(chan struct{}),
	}
}

// Start begins monitoring system resources
func (rm *ResourceMonitor) Start() {
	go rm.monitorLoop()
}

// Stop stops the resource monitor
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	if !rm.stopped {
		close(rm.stopCh)
		rm.stopped = true
	}
	rm.mu.Unlock()
}

func (rm *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(rm.updateInterval)
	defer ticker.Stop()

	// Initial update
	rm.updateMetrics()

	for {
		select {
		case <-ticker.C:
			rm.updateMetrics()
		case <-rm.stopCh:
			return
		}
	}
}

func (rm *ResourceMonitor) updateMetrics() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.previousMetrics = rm.currentMetrics

	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		rm.currentMetrics.CPUPercent = cpuPercent[0]
	}

	// Get memory usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		rm.currentMetrics.MemoryUsedMB = float64(memInfo.Used) / 1024 / 1024
		rm.currentMetrics.MemoryPercent = memInfo.UsedPercent
	}

	// Get disk IO stats
	ioCounters, err := disk.IOCounters()
	if err == nil {
		var totalReadBytes, totalWriteBytes, totalReadOps, totalWriteOps uint64
		for _, counter := range ioCounters {
			totalReadBytes += counter.ReadBytes
			totalWriteBytes += counter.WriteBytes
			totalReadOps += counter.ReadCount
			totalWriteOps += counter.WriteCount
		}
		rm.currentMetrics.IOReadBytes = totalReadBytes
		rm.currentMetrics.IOWriteBytes = totalWriteBytes
		rm.currentMetrics.IOReadOps = totalReadOps
		rm.currentMetrics.IOWriteOps = totalWriteOps
	}

	rm.currentMetrics.Timestamp = time.Now()
}

// GetMetrics returns the current system metrics
func (rm *ResourceMonitor) GetMetrics() SystemMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.currentMetrics
}

// CanRunJob checks if the system has enough resources to run a job
func (rm *ResourceMonitor) CanRunJob(constraints ResourceConstraints) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	metrics := rm.currentMetrics

	// Wait for metrics to be initialized
	if metrics.Timestamp.IsZero() {
		return false
	}

	// Check CPU constraint
	if constraints.MaxCPUPercent > 0 && metrics.CPUPercent >= constraints.MaxCPUPercent {
		return false
	}

	// Check memory constraint
	if constraints.MaxMemoryMB > 0 && metrics.MemoryUsedMB >= constraints.MaxMemoryMB {
		return false
	}

	// Check IO constraint (approximate ops per second)
	if constraints.MaxIOOpsPerSec > 0 {
		if rm.previousMetrics.Timestamp.IsZero() {
			return true // Not enough data yet
		}
		
		timeDiff := metrics.Timestamp.Sub(rm.previousMetrics.Timestamp).Seconds()
		if timeDiff > 0.1 { // Require at least 100ms of data
			readOpsDiff := metrics.IOReadOps - rm.previousMetrics.IOReadOps
			writeOpsDiff := metrics.IOWriteOps - rm.previousMetrics.IOWriteOps
			opsPerSec := int(float64(readOpsDiff+writeOpsDiff) / timeDiff)
			
			if opsPerSec >= constraints.MaxIOOpsPerSec {
				return false
			}
		}
	}

	return true
}

// GetGoRoutineCount returns the current number of goroutines
func GetGoRoutineCount() int {
	return runtime.NumGoroutine()
}

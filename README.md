# go-smart-queue

A smart Go job scheduler that runs tasks while monitoring CPU, memory, and IO usage. Queue jobs with constraints and priorities, using Go routines, channels, and OS metrics to throttle execution.

## Features

- ⚡ **Priority-based job scheduling** - Jobs are executed based on priority (higher values = higher priority)
- 📊 **Resource monitoring** - Real-time monitoring of CPU, memory, and IO usage
- 🎯 **Resource constraints** - Define maximum resource limits for jobs
- 🔄 **Concurrent execution** - Configurable worker pool using goroutines
- 🛑 **Job cancellation** - Cancel pending or running jobs
- 📈 **Real-time dashboard** - Live view of system metrics and job status
- 💻 **CLI interface** - Easy-to-use command-line interface

## Installation

```bash
git clone https://github.com/BaseMax/go-smart-queue.git
cd go-smart-queue
make build
```

Or using Go directly:

```bash
go build -o go-smart-queue
```

### Quick Start with Make

```bash
# Build the application
make build

# Run tests
make test

# Run the example
make example

# Clean build artifacts
make clean

# See all available commands
make help
```

## Usage

### Add a Job

Add a new job to the queue with specified constraints:

```bash
./go-smart-queue add --name "my-job" --priority 10 --max-cpu 80 --max-memory 1024 --sleep 5
```

Options:
- `--name, -n`: Job name (required)
- `--priority, -p`: Job priority (default: 5, higher = higher priority)
- `--max-cpu`: Maximum CPU percentage (default: 80.0)
- `--max-memory`: Maximum memory in MB (default: 1024.0)
- `--max-io`: Maximum IO ops per second (default: 1000)
- `--sleep, -s`: Sleep duration in seconds for demo (default: 5)
- `--command, -c`: Custom command to execute (optional)

### List Jobs

List all jobs (queued, running, and completed):

```bash
./go-smart-queue list
```

### View Status

Show system status and metrics:

```bash
./go-smart-queue status
```

Show status of a specific job:

```bash
./go-smart-queue status --job job-1
```

### Cancel a Job

Cancel a pending or running job:

```bash
./go-smart-queue cancel job-1
```

### Real-time Dashboard

Display a real-time dashboard with live system metrics and job status:

```bash
./go-smart-queue dashboard
```

Options:
- `--refresh, -r`: Refresh interval in seconds (default: 2)

## Architecture

### Core Components

1. **Job** - Represents a task with priority, constraints, and execution context
2. **ResourceMonitor** - Monitors system resources (CPU, memory, IO) in real-time
3. **Scheduler** - Manages job queue and worker pool
4. **Worker Pool** - Executes jobs concurrently using goroutines

### How It Works

1. Jobs are added to a priority queue
2. Resource monitor continuously tracks system metrics
3. Scheduler dispatches jobs to workers when:
   - Workers are available
   - System resources meet job constraints
4. Workers execute jobs concurrently using goroutines
5. Job status is tracked throughout its lifecycle

### Resource Throttling

Jobs are only executed when system resources are below their defined constraints:
- **CPU**: Job won't run if current CPU usage >= MaxCPUPercent
- **Memory**: Job won't run if current memory usage >= MaxMemoryMB
- **IO**: Job won't run if current IO ops/sec >= MaxIOOpsPerSec

## Examples

### High Priority Job

```bash
./go-smart-queue add --name "critical-backup" --priority 100 --max-cpu 90 --sleep 10
```

### Low Resource Job

```bash
./go-smart-queue add --name "email-task" --priority 5 --max-cpu 50 --max-memory 512 --sleep 3
```

### Multiple Jobs

```bash
# Add several jobs with different priorities
./go-smart-queue add --name "job1" --priority 10 --sleep 5
./go-smart-queue add --name "job2" --priority 20 --sleep 5
./go-smart-queue add --name "job3" --priority 5 --sleep 5

# View the dashboard
./go-smart-queue dashboard
```

### Programmatic Usage

See the `examples/` directory for a complete example of using the scheduler programmatically:

```bash
cd examples
./run-example.sh
```

The example demonstrates:
- Creating multiple jobs with different priorities
- Real-time monitoring of job execution
- Resource-based throttling
- Priority-based scheduling

## Testing

Run the test suite:

```bash
go test -v
# Or using make
make test

# With coverage report
make test-coverage
```

## Configuration

The scheduler is initialized with:
- **Max Workers**: 4 (configurable in main.go)
- **Resource Update Interval**: 1 second
- **Job Check Interval**: 500 milliseconds

## Development

### Project Structure

```
.
├── main.go           # CLI interface and main entry point
├── job.go            # Job structure and lifecycle
├── scheduler.go      # Scheduler and priority queue
├── monitor.go        # Resource monitoring
├── scheduler_test.go # Scheduler tests
├── monitor_test.go   # Monitor tests
├── go.mod            # Go module definition
└── README.md         # This file
```

### Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/shirou/gopsutil](https://github.com/shirou/gopsutil) - System metrics

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

BaseMax


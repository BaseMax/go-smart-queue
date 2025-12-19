# Examples

This directory contains example programs demonstrating the use of go-smart-queue.

## Running the Example

The example demonstrates using the job scheduler programmatically with multiple jobs of different priorities.

### Build and Run

```bash
# From the examples directory
go run example.go -I .. -tags example

# Or build from the root directory
cd ..
go run ./examples/example.go ./job.go ./monitor.go ./scheduler.go
```

Alternatively, you can copy the relevant source files:

```bash
cp ../job.go ../monitor.go ../scheduler.go .
go run *.go
```

## Demo Script

The `demo.sh` script demonstrates the CLI functionality:

```bash
./demo.sh
```

Note: The demo script requires building the main CLI application first from the root directory.

## Example Output

The example will:
1. Create 5 jobs with different priorities
2. Monitor their execution in real-time
3. Display system metrics (CPU, memory, IO)
4. Show final results with job completion status and duration

You should see output similar to:

```
Starting job scheduler example...

Added job: Data Processing (ID: job-1, Priority: 20)
Added job: Email Notifications (ID: job-2, Priority: 15)
...

Jobs added to queue. Monitoring execution...

  → Started: Processing data
  → Started: Generating reports
[20:11:20] Stats: Queued=3, Active=2, Completed=0 | CPU=0.0%, Mem=1519MB
...

All jobs completed!
```

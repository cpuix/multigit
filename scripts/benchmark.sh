#!/bin/bash

set -e

# Create output directory
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_DIR="profiles/${TIMESTAMP}"
mkdir -p "${OUTPUT_DIR}"

# Build the application
echo "Building the application..."
go build -o multigit .

# Run benchmarks with profiling
echo "Running benchmarks with profiling..."

# CPU and Memory profiling
GODEBUG=gctrace=1 \
GOMAXPROCS=1 \
  go test -v \
  -bench=. \
  -benchmem \
  -benchtime=5s \
  -cpuprofile "${OUTPUT_DIR}/cpu.prof" \
  -memprofile "${OUTPUT_DIR}/mem.prof" \
  -blockprofile "${OUTPUT_DIR}/block.prof" \
  -mutexprofile "${OUTPUT_DIR}/mutex.prof" \
  ./internal/ssh/...

# Generate flamegraphs
if command -v go-torch &> /dev/null; then
  echo "Generating flamegraphs..."
  
  # CPU flamegraph
  go-torch "${OUTPUT_DIR}/cpu.prof" -o "${OUTPUT_DIR}/flamegraph_cpu.svg"
  
  # Memory flamegraph
  go-torch --inuse_space "${OUTPUT_DIR}/mem.prof" -o "${OUTPUT_DIR}/flamegraph_mem_inuse.svg"
  go-torch --alloc_space "${OUTPUT_DIR}/mem.prof" -o "${OUTPUT_DIR}/flamegraph_mem_alloc.svg"
  
  # Block flamegraph
  go-torch "${OUTPUT_DIR}/block.prof" -o "${OUTPUT_DIR}/flamegraph_block.svg"
  
  # Mutex flamegraph
  go-torch "${OUTPUT_DIR}/mutex.prof" -o "${OUTPUT_DIR}/flamegraph_mutex.svg"
fi

# Generate pprof text output
echo "Generating pprof text output..."
go tool pprof -text "${OUTPUT_DIR}/cpu.prof" > "${OUTPUT_DIR}/cpu.txt"
go tool pprof -text "${OUTPUT_DIR}/mem.prof" > "${OUTPUT_DIR}/mem.txt"

# Generate memory allocation by function
go tool pprof -alloc_space -text "${OUTPUT_DIR}/mem.prof" > "${OUTPUT_DIR}/mem_alloc.txt"

# Generate memory in-use by function
go tool pprof -inuse_space -text "${OUTPUT_DIR}/mem.prof" > "${OUTPUT_DIR}/mem_inuse.txt"

echo "Benchmark results and profiles saved to: ${OUTPUT_DIR}"
echo "You can analyze the results using:"
echo "  go tool pprof -http=:8080 ${OUTPUT_DIR}/cpu.prof"
echo "  go tool pprof -http=:8080 ${OUTPUT_DIR}/mem.prof"

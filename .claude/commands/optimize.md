---
allowed-tools: Task(mage-x-benchmark), Task(mage-x-analyzer), Task(mage-x-refactor), Task(mage-x-builder), Bash(go test -bench:*), Bash(go build:*), Read, Write, MultiEdit, Grep, Glob
argument-hint: [performance|memory|build|all]
description: Performance analysis and optimization with benchmarking
model: claude-sonnet-4-20250514
---

## Context
- Benchmarks available: !`find . -name "*_test.go" -exec grep -l "^func Benchmark" {} \; | wc -l | xargs echo "Benchmark files:"`
- Build time: !`time go build ./... 2>&1 | grep real | awk '{print $2}' || echo "Not measured"`

## Your Task

Optimize code for better performance through parallel analysis. Focus: ${ARGUMENTS:-performance}

### Phase 1: Parallel Performance Analysis

1. **mage-x-benchmark** - Performance Profiling:
   - Run existing benchmarks
   - Identify performance hotspots
   - Memory allocation analysis
   - CPU profiling
   - Goroutine analysis

2. **mage-x-analyzer** - Code Analysis:
   - Algorithm complexity
   - Inefficient patterns
   - Resource usage
   - Concurrency issues
   - Cache efficiency

3. **mage-x-builder** - Build Optimization:
   - Compilation time analysis
   - Binary size optimization
   - Build flag improvements
   - Dependency analysis

### Phase 2: Optimization Implementation

1. **mage-x-refactor** - Performance Improvements:
   - Algorithm optimization
   - Memory allocation reduction
   - Concurrency enhancements
   - Cache-friendly refactoring
   - Build optimizations

2. **Validation**:
   - Re-run benchmarks
   - Compare before/after metrics
   - Ensure functionality preserved
   - Document improvements

## Optimization Focus Areas

- **performance**: CPU and execution speed
- **memory**: Memory usage and allocations
- **build**: Compilation time and binary size
- **all**: Comprehensive optimization

## Expected Output

### Performance Analysis
```
Current Performance Metrics:
- Execution Time: [baseline]
- Memory Usage: [baseline]
- Allocations: [count]
- Build Time: [duration]
```

### Identified Bottlenecks
1. **Hot Paths**: Functions consuming most CPU
2. **Memory Hogs**: High allocation areas
3. **Inefficient Algorithms**: O(n²) or worse
4. **Concurrency Issues**: Lock contention

### Applied Optimizations

#### Algorithm Improvements
```go
// Before: O(n²) nested loops
// After: O(n log n) optimized algorithm
```

#### Memory Optimizations
```go
// Before: Multiple allocations
// After: Reused buffers
```

#### Concurrency Enhancements
```go
// Before: Sequential processing
// After: Parallel execution with worker pool
```

### Performance Gains
```
Improvement Summary:
- Execution: -45% (2.3s → 1.3s)
- Memory: -60% (120MB → 48MB)
- Allocations: -70% (1M → 300K)
- Build Time: -20% (30s → 24s)
```

### Benchmark Results
```
BenchmarkBefore-8    1000    1,234,567 ns/op    45678 B/op    123 allocs/op
BenchmarkAfter-8     3000      456,789 ns/op    12345 B/op     45 allocs/op
```

# Go Performance Reference

## Principle: measure first

Never optimise without a profile. Gut-feel optimisation wastes time and introduces bugs. The process is always:

1. Reproduce the slow path with a benchmark or realistic workload
2. Profile with `pprof`
3. Identify the actual bottleneck
4. Fix the bottleneck
5. Measure again to confirm improvement

## pprof — CPU profiling

### In tests / benchmarks

```bash
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof -http=:8080 cpu.prof
```

### In a running service

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Endpoints: `http://localhost:6060/debug/pprof/`

```bash
# 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Heap profiling — find allocations

```bash
go test -bench=. -memprofile=mem.prof ./...
go tool pprof -http=:8080 -alloc_space mem.prof
```

Look for the top allocating functions. Many allocations create GC pressure even if individual objects are small.

## Escape analysis

The compiler moves variables to the heap when it can't prove they're stack-safe (e.g. their address escapes). Heap allocations cost more than stack allocations.

```bash
go build -gcflags="-m" ./...
```

Output lines like `moved to heap` tell you what escaped. Common escapes:

- Returning a pointer to a local variable
- Storing a value in an interface
- Closing over a variable in a goroutine

Escapes aren't always bad — avoid premature optimisation here — but high-frequency hot-path escapes are worth addressing.

## Reducing allocations — common wins

### Preallocate slices when length is known

```go
// Bad: grows the backing array repeatedly
var results []User
for _, id := range ids {
    results = append(results, fetch(id))
}

// Good: single allocation
results := make([]User, 0, len(ids))
for _, id := range ids {
    results = append(results, fetch(id))
}
```

### strings.Builder for string concatenation in loops

```go
// Bad: allocates a new string on every +
s := ""
for _, part := range parts {
    s += part
}

// Good: single allocation
var b strings.Builder
b.Grow(estimatedSize)
for _, part := range parts {
    b.WriteString(part)
}
result := b.String()
```

### sync.Pool for short-lived objects

Reuse allocations of frequently created objects (e.g. buffers, request structs) across goroutines:

```go
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    // use buf
    return buf.String()
}
```

`sync.Pool` objects may be collected at any GC cycle — don't use it for anything that must persist.

### Avoid interface boxing in hot paths

Storing a concrete type in an `interface{}` / `any` allocates on the heap if the type doesn't fit in a pointer. Measure before worrying about this.

## I/O — buffering

Unbuffered I/O is slow. Always buffer when reading or writing in a loop:

```go
// Reading
f, _ := os.Open("data.txt")
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    process(scanner.Bytes())
}

// Writing
f, _ := os.Create("out.txt")
bw := bufio.NewWriter(f)
defer bw.Flush() // don't forget
for _, line := range lines {
    fmt.Fprintln(bw, line)
}
```

## Goroutine overhead

Goroutines start at ~2–8 KB stack and grow as needed. Spawning millions is expensive. Use a worker pool (see `concurrency.md`) when processing large numbers of items.

## Benchmarking correctly

```go
func BenchmarkEncode(b *testing.B) {
    payload := generatePayload() // outside the loop
    b.ResetTimer()               // don't count setup time
    b.ReportAllocs()             // show allocs/op

    for b.Loop() {               // Go 1.24+; use i := 0; i < b.N; i++ in older versions
        _ = encode(payload)
    }
}
```

Always run benchmarks at least twice and compare with `benchstat`:

```bash
go test -bench=. -count=5 ./... | tee new.txt
benchstat old.txt new.txt
```

## Memory leak patterns

- **Goroutine leak**: goroutine blocked on channel send/receive with no consumer. Always ensure goroutines can exit.
- **Timer leak**: `time.After` in a loop creates a new timer per iteration that lives until it fires. Use `time.NewTimer` and reset.
- **Slice backing array retention**: slicing a large slice keeps the whole backing array alive. Copy if you only need a small portion.

```go
// Retains the large backing array
small := big[0:3]

// Releases the large backing array
small := append([]byte(nil), big[0:3]...)
```

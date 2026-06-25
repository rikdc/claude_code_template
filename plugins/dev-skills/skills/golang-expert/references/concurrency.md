# Go Concurrency Reference

## Core principle

A goroutine is a liability until you can answer: **when does it exit and who knows when it's done?** If you can't answer that, the goroutine is a leak waiting to happen.

## Context cancellation — the primary lifecycle tool

```go
func process(ctx context.Context, jobs <-chan Job) error {
    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("process cancelled: %w", ctx.Err())
        case job, ok := <-jobs:
            if !ok {
                return nil // channel closed cleanly
            }
            if err := handle(ctx, job); err != nil {
                return fmt.Errorf("handle job %s: %w", job.ID, err)
            }
        }
    }
}
```

## errgroup — fan-out with error collection

`golang.org/x/sync/errgroup` is the standard way to run N goroutines and collect the first error:

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, ids []string) ([]*Item, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([]*Item, len(ids))

    for i, id := range ids {
        i, id := i, id // Go <1.22: capture loop vars
        g.Go(func() error {
            item, err := fetch(ctx, id)
            if err != nil {
                return fmt.Errorf("fetch %s: %w", id, err)
            }
            results[i] = item
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

Use `g.SetLimit(n)` to bound concurrency when the slice is large.

## Worker pool — bounded concurrency

```go
type Pool struct {
    work chan func()
    wg   sync.WaitGroup
}

func NewPool(workers int) *Pool {
    p := &Pool{work: make(chan func(), workers*2)}
    for range workers {
        p.wg.Add(1)
        go func() {
            defer p.wg.Done()
            for fn := range p.work {
                fn()
            }
        }()
    }
    return p
}

func (p *Pool) Submit(fn func()) { p.work <- fn }
func (p *Pool) Close()          { close(p.work); p.wg.Wait() }
```

## Pipeline pattern

```go
func generate(ctx context.Context, vals ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, v := range vals {
            select {
            case out <- v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

Always close output channels in the sending goroutine. Never close from the receiver.

## sync primitives

### Mutex — protect shared state

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

Prefer `sync.RWMutex` only when reads significantly outnumber writes and profiling shows the mutex is a bottleneck.

### Once — one-time initialisation

```go
var (
    instance *Client
    once     sync.Once
)

func GetClient() *Client {
    once.Do(func() { instance = newClient() })
    return instance
}
```

### WaitGroup — wait for a known number of goroutines

```go
var wg sync.WaitGroup
for _, task := range tasks {
    wg.Add(1)
    go func(t Task) {
        defer wg.Done()
        run(t)
    }(task)
}
wg.Wait()
```

`Add` must be called before `go`. Never call `Add` inside the goroutine.

## Common pitfalls

### Loop variable capture (pre-Go 1.22)

```go
// BUG: all goroutines share the same `v`
for _, v := range items {
    go func() { process(v) }()
}

// Fix: capture explicitly
for _, v := range items {
    v := v
    go func() { process(v) }()
}
// Go 1.22+: loop variables are per-iteration — no capture needed
```

### Nil channel blocks forever

```go
var ch chan int  // nil
ch <- 1          // blocks forever
<-ch             // blocks forever
// Use select with default or always initialise channels with make
```

### Goroutine leak via unbuffered channel

```go
// Leaks if no one reads result
func leak() {
    ch := make(chan int)
    go func() { ch <- compute() }() // stuck if caller returns
}

// Fix: buffer of 1 so goroutine can always send
ch := make(chan int, 1)
```

## Race detector

Always run `go test -race ./...` before considering concurrent code correct. The race detector catches real bugs; false positives are extremely rare.

Add to CI:

```makefile
test:
	go test -race -count=1 ./...
```

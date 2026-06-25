# Go Error Handling Reference

## The single handling rule

An error is either **returned** to the caller or **logged** at the boundary where it stops propagating — never both. Logging and returning causes duplicate log lines and obscures where an error actually originated.

```go
// WRONG: log and return
if err != nil {
    log.Error("failed to save", "err", err)
    return fmt.Errorf("save: %w", err)
}

// RIGHT at intermediate layer: return with context
if err != nil {
    return fmt.Errorf("save user %s: %w", id, err)
}

// RIGHT at boundary (HTTP handler, main, job runner): log and stop
if err != nil {
    slog.Error("save user", "id", id, "err", err)
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
}
```

## Wrapping rules

- Use `fmt.Errorf("context: %w", err)` — `%w` preserves the chain for `errors.Is` / `errors.As`
- Use `%v` (not `%w`) only when you intentionally want to break the chain (rare — e.g. exposing to external callers where internal types should not leak)
- Error strings: lowercase, no trailing punctuation — they get concatenated: `"parse config: open file: no such file"`

```go
// Good
return fmt.Errorf("open config %s: %w", path, err)

// Bad — breaks chain
return fmt.Errorf("open config %s: %v", path, err)

// Bad — uppercase / punctuation
return fmt.Errorf("Failed to open config: %w", err)
```

## Sentinel errors

Define at package level for expected conditions that callers need to distinguish:

```go
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrUnauthorised  = errors.New("unauthorised")
)

// Check with errors.Is — works through wrapping chains
if errors.Is(err, ErrNotFound) {
    http.Error(w, "resource not found", http.StatusNotFound)
    return
}
```

Sentinel errors are cheap (allocated once), comparable, and chain-safe.

## Custom error types

Use when callers need structured data from the error, not just identity:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

// Extract with errors.As
var ve *ValidationError
if errors.As(err, &ve) {
    http.Error(w, ve.Message, http.StatusBadRequest)
    return
}
```

## errors.Join — combining independent errors

```go
// Go 1.20+
var errs []error
for _, item := range items {
    if err := validate(item); err != nil {
        errs = append(errs, err)
    }
}
if err := errors.Join(errs...); err != nil {
    return err
}
```

## Structured logging with slog

Use `slog` (Go 1.21+) at error boundaries. Attach structured fields, not interpolated strings:

```go
slog.Error("process payment",
    "payment_id", payment.ID,
    "amount",     payment.Amount,
    "err",        err,
)
```

Never include PII (email, card number, tokens) in log messages. Log IDs, then look up details separately.

## panic / recover

`panic` is for **truly unrecoverable states** — programmer errors, impossible conditions, init failures. Never use it for expected runtime errors.

```go
// Acceptable: invariant violation in a constructor
func NewServer(cfg Config) *Server {
    if cfg.Port == 0 {
        panic("server port must be non-zero")
    }
    return &Server{cfg: cfg}
}

// Always recover at goroutine boundaries to prevent process crash
go func() {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("goroutine panicked", "panic", r)
        }
    }()
    doWork()
}()
```

## HTTP error translation

Internal errors must never leak to clients. Translate at the handler boundary:

```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user, err := svc.GetUser(r.Context(), id)
    if err != nil {
        switch {
        case errors.Is(err, ErrNotFound):
            http.Error(w, "user not found", http.StatusNotFound)
        case errors.Is(err, ErrUnauthorised):
            http.Error(w, "forbidden", http.StatusForbidden)
        default:
            slog.Error("get user", "id", id, "err", err)
            http.Error(w, "internal server error", http.StatusInternalServerError)
        }
        return
    }
    writeJSON(w, http.StatusOK, user)
}
```

## Decision table

| Situation | Use |
|-----------|-----|
| Distinguishable condition, no data | Sentinel (`errors.New`) |
| Condition with structured data | Custom type + `errors.As` |
| Intermediate propagation | `fmt.Errorf("context: %w", err)` |
| Multiple independent failures | `errors.Join` |
| Unrecoverable programmer error | `panic` |
| Log-and-stop at boundary | `slog.Error` + return/abort |

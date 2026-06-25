# Go Project Structure Reference

## Standard layout for a service

```
myservice/
├── cmd/
│   └── myservice/
│       └── main.go          # Entrypoint: wire dependencies, start server
├── internal/
│   ├── handler/             # HTTP/gRPC handlers — decode input, call service, encode response
│   ├── service/             # Business logic — orchestrates domain operations
│   ├── repository/          # Data access — SQL, cache, external APIs
│   ├── domain/              # Core types and interfaces (no external dependencies)
│   └── middleware/          # HTTP middleware — auth, logging, tracing
├── pkg/                     # Packages safe for external import (think carefully before adding)
├── migrations/              # SQL migration files
├── Makefile
├── go.mod
└── go.sum
```

`internal/` enforces the boundary: nothing outside the module can import it. Prefer `internal/` for everything by default.

## Package design rules

- **One concept per package** — a package named `util` or `common` is a design smell
- **Package name = directory name** — no `package userutil` in a directory called `user`
- **Import aliases lowercase** — `import flog "github.com/org/flog"`, never `fLog`
- **Define interfaces in the consumer** — the package that *uses* `Store` should declare the `Store` interface, not the package that implements it
- **No circular imports** — if you have one, the package boundary is wrong

## go.mod essentials

```
module github.com/org/myservice

go 1.26

require (
    golang.org/x/sync v0.7.0
)
```

- Commit both `go.mod` and `go.sum`
- Run `go mod tidy` after adding or removing imports
- Use `replace` directives only temporarily (local development); remove before merging

## Clean architecture layer rules

| Layer | May import | Must not import |
|-------|-----------|----------------|
| `domain/` | stdlib only | anything internal |
| `repository/` | `domain/`, stdlib, drivers | `handler/`, `service/` |
| `service/` | `domain/`, `repository/` interfaces | `handler/`, infrastructure |
| `handler/` | `service/` interfaces, `domain/` | `repository/` directly |
| `main.go` | everything | — |

Dependencies point inward. `handler` never reaches into `repository` directly.

## Dependency injection in main.go

Wire everything in `main.go` — no `init()` functions for side effects, no globals for dependencies:

```go
func main() {
    cfg := config.Load()

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        slog.Error("open database", "err", err)
        os.Exit(1)
    }

    repo    := repository.NewUserRepo(db)
    svc     := service.NewUserService(repo, slog.Default())
    handler := handler.NewUserHandler(svc)

    mux := http.NewServeMux()
    mux.HandleFunc("GET /users/{id}", handler.GetUser)

    srv := &http.Server{Addr: cfg.Addr, Handler: mux}
    // ... start and graceful shutdown
}
```

## Naming conventions

- **Types**: PascalCase nouns — `UserService`, `PaymentRepo`
- **Interfaces**: adjective or verb phrase — `Storer`, `UserGetter`; avoid `IUserService` prefixes (legacy pattern)
- **Functions**: verb phrase — `GetUser`, `CreatePayment`
- **Booleans**: `IsActive`, `HasPermission`, `CanRetry`
- **Errors**: `ErrNotFound`, `ErrAlreadyExists`
- **Constants**: PascalCase, not ALL_CAPS
- **Initialisations**: same case — `httpURL`, `userID`, `apiToken`

## Configuration

Load config at startup from environment variables. Never from flags inside business logic. Never hardcode:

```go
type Config struct {
    Addr        string
    DatabaseURL string
    LogLevel    slog.Level
}

func Load() Config {
    return Config{
        Addr:        env("ADDR", ":8080"),
        DatabaseURL: mustEnv("DATABASE_URL"),
        LogLevel:    parseLogLevel(env("LOG_LEVEL", "info")),
    }
}
```

## Multi-binary repos

When one module contains multiple binaries, each gets its own directory under `cmd/`:

```
cmd/
├── api/main.go
├── worker/main.go
└── migrate/main.go
```

Build individually: `go build ./cmd/api`

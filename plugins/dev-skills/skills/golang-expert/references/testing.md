# Go Testing Reference

## Table-driven tests — the standard pattern

```go
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name     string
        price    float64
        discount float64
        want     float64
        wantErr  bool
    }{
        {name: "zero discount",    price: 100, discount: 0,   want: 100},
        {name: "half off",         price: 100, discount: 0.5, want: 50},
        {name: "full price",       price: 50,  discount: 1.0, want: 0},
        {name: "negative price",   price: -1,  discount: 0.1, wantErr: true},
        {name: "discount over 1",  price: 100, discount: 1.1, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateDiscount(tt.price, tt.discount)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.InDelta(t, tt.want, got, 0.001)
        })
    }
}
```

Use `testify/require` for assertions that must stop the test on failure, `testify/assert` for non-fatal checks.

## Mocking interfaces

Define interfaces at the consumer; mock them in tests. Use `github.com/stretchr/testify/mock`:

```go
// production interface (in handler or service package)
type UserStore interface {
    Get(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, u *User) error
}

// mock (in _test.go or testutil package)
type mockUserStore struct {
    mock.Mock
}

func (m *mockUserStore) Get(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// test usage
func TestUserService_GetUser(t *testing.T) {
    store := new(mockUserStore)
    store.On("Get", mock.Anything, "u1").Return(&User{ID: "u1"}, nil)

    svc := NewUserService(store)
    user, err := svc.GetUser(context.Background(), "u1")
    require.NoError(t, err)
    assert.Equal(t, "u1", user.ID)
    store.AssertExpectations(t)
}
```

For simpler cases, hand-roll a stub — a mock framework is overhead if you only need one method.

## Testing HTTP handlers

Use `net/http/httptest` — no real server needed:

```go
func TestGetUserHandler(t *testing.T) {
    store := new(mockUserStore)
    store.On("Get", mock.Anything, "42").Return(&User{ID: "42", Name: "Alice"}, nil)

    h := NewUserHandler(store)
    rec := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
    req.SetPathValue("id", "42") // Go 1.22+

    h.GetUser(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
    var got User
    require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
    assert.Equal(t, "Alice", got.Name)
}
```

## Race detector — mandatory

```bash
go test -race ./...
```

Add to Makefile and CI. A test suite without `-race` provides false confidence for concurrent code.

## Benchmarks

```go
func BenchmarkParseConfig(b *testing.B) {
    data := []byte(`{"port":8080,"timeout":"30s"}`)
    b.ResetTimer()
    for b.Loop() { // Go 1.24+; use b.N in older versions
        _, _ = ParseConfig(data)
    }
}
```

Run: `go test -bench=. -benchmem ./...`

`-benchmem` shows allocations per operation. A reduction in `allocs/op` often matters more than ns/op.

## Fuzz testing

Add fuzz targets for any function that parses untrusted input:

```go
func FuzzParseID(f *testing.F) {
    f.Add("user-123")
    f.Add("")
    f.Add("../../../../etc/passwd")

    f.Fuzz(func(t *testing.T, input string) {
        _, _ = ParseID(input) // must not panic
    })
}
```

Run: `go test -fuzz=FuzzParseID -fuzztime=30s`

Commit the seed corpus in `testdata/fuzz/FuzzParseID/`.

## Integration tests

Tag integration tests so they don't run in unit-test mode:

```go
//go:build integration

package repository_test

func TestUserRepo_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // use a real database
}
```

Run: `go test -tags=integration ./...`

## Test helpers

Keep test helpers in `testutil/` or `internal/testutil/`. Helpers that call `t.Fatal` must accept `testing.TB` (works for both `*testing.T` and `*testing.B`):

```go
func MustOpenDB(t testing.TB, dsn string) *sql.DB {
    t.Helper()
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("open db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

`t.Helper()` marks the function so test failure lines point to the caller, not the helper.

## What not to test

- Internal unexported functions (test through the public API)
- `main()` directly (extract logic into testable functions)
- Third-party libraries
- Framework glue code with no business logic

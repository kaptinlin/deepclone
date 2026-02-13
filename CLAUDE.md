# Project Guidelines

## Overview

`github.com/kaptinlin/deepclone` is a high-performance deep cloning library for Go. It provides a single generic function `Clone[T any](src T) T` that creates a fully independent deep copy of any Go value, with automatic circular reference detection.

- **Go version**: 1.26
- **Runtime dependencies**: None (standard library only)
- **Test dependency**: `github.com/stretchr/testify`

## Build Commands

```bash
make test             # Run all tests with -race
make test-coverage    # Tests with coverage report (coverage.html)
make test-verbose     # Verbose test output with -race
make lint             # golangci-lint + go mod tidy check
make bench            # Run benchmarks
make bench-comparison # Run comparison benchmarks against other libraries
make fmt              # Format Go code
make vet              # Run go vet
make verify           # Full pipeline: deps + fmt + vet + lint + test
```

## Project Structure

```
clone.go              # Main implementation: Clone function and all cloning logic
types.go              # Cloneable interface definition
doc.go                # Package-level GoDoc documentation
clone_test.go         # Unit tests (~800 lines, 95%+ coverage)
cache_test.go         # Struct cache behavior tests
concurrent_test.go    # Concurrent stress tests (100+ goroutines)
benchmark_test.go     # Benchmarks using testing.B.Loop() (Go 1.24+)
example_test.go       # Testable examples for GoDoc
examples/             # Runnable example programs (basic, circular, custom)
benchmarks/           # Separate module for comparison benchmarks against 4 libraries
```

## Architecture

### Public API

There are only three exported symbols:

- `Clone[T any](src T) T` -- Deep clone any value
- `Cloneable` interface -- Implement `Clone() any` for custom cloning behavior
- `CacheStats() (entries, fields int)` / `ResetCache()` -- Struct cache observability and control

### Cloning Strategy (Hierarchical)

Clone uses a tiered approach, checked in order:

1. **Ultra-fast path**: Primitive types (`int`, `string`, `bool`, `float`, `complex`, `uintptr`) return as-is with zero allocation.
2. **Fast path - slices**: Common slice types (`[]int`, `[]string`, `[]byte`, `[]float64`, etc.) use a generic `cloneSliceExact` helper that preserves capacity.
3. **Fast path - maps**: Common map types with primitive keys/values (`map[string]int`, `map[int]string`, etc.) use `maps.Clone()`. `map[string]any` is deliberately excluded to handle potential circular references via reflection.
4. **Cloneable interface**: Types implementing `Cloneable` delegate to their `Clone()` method. Circular reference detection does not apply inside custom Clone methods.
5. **Reflection path**: All remaining types use cached reflection with full circular reference detection.

### Struct Cache

`structCache` maps `reflect.Type` to `structTypeInfo` (per-field clone action decisions). Protected by `sync.RWMutex` with double-check locking. The cache is bounded by the number of distinct struct types in the program (compile-time determined), so LRU eviction is unnecessary.

### Circular Reference Detection

`cloneContext.visited` maps `uintptr` to `reflect.Value`. Pre-allocated with capacity 8. Pointers are stored in the visited map before recursing into the pointed value, so self-referencing structures resolve correctly.

For slices, only those with element kinds that can contain cycles (pointer, interface, slice, map, struct) are tracked. Cached slices are validated for matching length/capacity to prevent aliasing bugs.

### Special Cases

- **Channels**: Return zero value of same type
- **Functions**: Return as-is (cannot be deep cloned)
- **Nil pointers/slices/maps**: Return as-is
- **Unexported struct fields**: Skipped (cannot access via reflection)
- **Type aliases**: Handled via `ConvertibleTo`/`AssignableTo` checks in map and struct cloning

## Code Conventions

- All tests must pass with `-race` flag
- Benchmarks use `b.Loop()` (Go 1.24+), not `for i := 0; i < b.N; i++`
- Use `for range N` syntax (Go 1.22+) for integer loops
- Use `any` instead of `interface{}`
- Use `reflect.Pointer` instead of deprecated `reflect.Ptr`
- Test assertions use `testify/assert` and `testify/require`
- No external runtime dependencies -- standard library only for core logic

## Linting

- golangci-lint version managed via `.golangci.version` file (currently 2.9.0)
- 21 linters enabled (see `.golangci.yml`)
- Test files exclude `gosec`, `noctx`, `revive`
- All issues reported (no max-issues caps)

## Testing Patterns

- Table-driven tests with `t.Run()` subtests
- Memory address assertions (`assert.NotSame`) for independence verification
- `t.Cleanup(ResetCache)` between cache tests
- Concurrent tests: 100-200 goroutines, 100-500 iterations each
- Coverage target: 95%+

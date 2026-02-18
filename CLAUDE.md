# DeepClone

High-performance deep cloning library for Go with zero-allocation fast paths and automatic circular reference detection.

**Module**: `github.com/kaptinlin/deepclone`
**Go Version**: 1.26
**Dependencies**: None (stdlib only, testify for tests)

## Commands

```bash
task test              # Run tests with -race
task test-coverage     # Generate coverage.html report
task test-verbose      # Verbose test output with -race
make bench             # Run benchmarks
make bench-comparison  # Compare against other libraries (benchmarks/ module)
task lint              # golangci-lint + go mod tidy check
make fmt               # Format code
make vet               # Run go vet
task verify            # Full pipeline: deps + fmt + vet + lint + test
```

## Architecture

Single-file library (`clone.go`) with minimal surface area:

```
clone.go              # All cloning logic (432 lines)
types.go              # Cloneable interface
doc.go                # Package documentation
*_test.go             # Tests (800+ lines, 95%+ coverage)
examples/             # Runnable examples
benchmarks/           # Separate comparison module
```

### Public API (3 Exports Only)

```go
func Clone[T any](src T) T                    // Deep clone any value
type Cloneable interface { Clone() any }      // Custom cloning behavior
func CacheStats() (entries, fields int)       // Cache observability
func ResetCache()                             // Cache control (tests only)
```

### Hierarchical Cloning Strategy

Clone checks these paths in order (clone.go:142-247):

1. **Ultra-fast path** (lines 144-150): Primitives (`int`, `string`, `bool`, etc.) return as-is with zero allocation
2. **Fast path - slices** (lines 153-182): Common types (`[]int`, `[]string`, `[]byte`, etc.) use `cloneSliceExact[S, E]` generic helper that preserves capacity
3. **Fast path - maps** (lines 186-222): Simple maps (`map[string]int`, `map[int]string`, etc.) use `maps.Clone()`. Excludes `map[string]any` to handle circular references via reflection
4. **Cloneable interface** (lines 230-234): Types implementing `Cloneable` delegate to their `Clone()` method. Circular reference detection does NOT apply inside custom Clone methods
5. **Reflection path** (lines 240-246): All other types use cached reflection with full circular reference detection

### Struct Cache (clone.go:16-111)

- `structCache` maps `reflect.Type` → `structTypeInfo` (per-field clone actions)
- Protected by `sync.RWMutex` with double-check locking (lines 42-55)
- Bounded by number of distinct struct types (compile-time determined), no LRU needed
- Each field marked as `copyField` (primitives, unexported) or `cloneField` (complex types)

### Circular Reference Detection (clone.go:24-33, 280-301)

- `cloneContext.visited` maps `uintptr` → `reflect.Value`
- Pre-allocated with capacity 8
- Pointers stored in visited map BEFORE recursing (line 293) to handle self-references
- Slices tracked only if element kind can contain cycles (pointer, interface, slice, map, struct)
- Cached slices validated for matching length/capacity to prevent aliasing bugs (lines 316-319)

## Key Types and Interfaces

### Cloneable Interface (types.go:28-30)

```go
type Cloneable interface {
    Clone() any
}
```

Implement for custom cloning behavior. The library's circular reference detection does NOT apply inside your Clone method - handle cycles manually if needed.

### Internal Types

- `cloneContext`: Tracks visited objects during reflection-based cloning
- `structTypeInfo`: Cached per-field clone actions for struct types
- `fieldAction`: Enum (`copyField` or `cloneField`)

## Coding Rules

### Go 1.26 Features (Mandatory)

- Use `for range N` instead of `for i := 0; i < N; i++` (clone.go:328, 408)
- Use `b.Loop()` in benchmarks instead of `for i := 0; i < b.N; i++` (benchmark_test.go)
- Use `reflect.MakeMapWithSize(type, size)` for map initialization (clone.go:348)
- Use `clear(map)` for map clearing (clone.go:110)

### Conventions

- All tests MUST pass with `-race` flag
- Use `any` instead of `interface{}`
- Use `reflect.Pointer` instead of deprecated `reflect.Ptr`
- Test assertions use `testify/assert` and `testify/require`
- Memory independence verified with `assert.NotSame(t, &original, &cloned)`
- No external runtime dependencies - stdlib only for core logic

### Special Cases (clone.go:268-278)

- **Channels**: Return zero value of same type
- **Functions**: Return as-is (cannot be deep cloned)
- **Nil pointers/slices/maps**: Return as-is
- **Unexported struct fields**: Skipped (cannot access via reflection)
- **Type aliases**: Handled via `ConvertibleTo`/`AssignableTo` checks (clone.go:361-366, 394-396)

## Testing

### Test Structure

- Table-driven tests with `t.Run()` subtests
- Concurrent tests: 100-200 goroutines, 100-500 iterations each (concurrent_test.go)
- Cache tests use `t.Cleanup(ResetCache)` for isolation (cache_test.go)
- Coverage target: 95%+

### Test Files

```
clone_test.go         # Core cloning tests (800+ lines)
cache_test.go         # Struct cache behavior
concurrent_test.go    # Concurrent stress tests
benchmark_test.go     # Performance benchmarks
example_test.go       # Testable examples for GoDoc
```

### Running Tests

```bash
task test              # Standard test run with -race
task test-coverage     # Generate coverage.html
task test-verbose      # Verbose output
make bench             # Run benchmarks
```

## Performance

### Optimization Principles

- **Zero allocations** for primitive types (verified by benchmarks)
- **Fast paths** for common slice/map types avoid reflection overhead
- **Reflection caching** for struct types (computed once per type)
- **Capacity preservation** for slices and maps (clone.go:118, 348)
- **Minimal overhead** for complex operations

### Benchmark Results (Apple M3, darwin/arm64)

| Operation | Performance | Memory | Allocations |
|-----------|-------------|---------|-------------|
| Primitives | 2.7-3.6 ns/op | 0 B/op | 0 allocs/op |
| Slice (100 ints) | 200.6 ns/op | 896 B/op | 1 allocs/op |
| Map (100 entries) | 4,299 ns/op | 3,544 B/op | 4 allocs/op |
| Simple Struct | 248.6 ns/op | 128 B/op | 4 allocs/op |
| Nested Struct | 1,386 ns/op | 952 B/op | 19 allocs/op |
| Large Slice (10K) | 6,709 ns/op | 81,920 B/op | 1 allocs/op |

Run `make bench-comparison` for detailed comparisons with other libraries.

## Error Handling

This library does not return errors. Invalid operations result in:

- **Invalid values**: Return as-is
- **Nil values**: Return as-is
- **Channels**: Return zero value
- **Unexported fields**: Silently skipped
- **Type conversion failures**: Field skipped (maps/structs)

## Linting

- **Version**: Managed via `.golangci.version` (currently 2.9.0)
- **Config**: `.golangci.yml` with 21 linters enabled
- **Test exclusions**: `gosec`, `noctx`, `revive` disabled for `*_test.go`
- **Issues**: No max-issues caps (all issues reported)

Run `task lint` to execute golangci-lint and go mod tidy checks.


## Agent Skills

This package indexes agent skills from its own .agents/skills directory (deepclone/.agents/skills/):

| Skill | When to Use |
|-------|-------------|
| [agent-md-creating](.agents/skills/agent-md-creating/) | Create or update CLAUDE.md and AGENTS.md instructions for this Go package. |
| [code-simplifying](.agents/skills/code-simplifying/) | Refine recently changed Go code for clarity and consistency without behavior changes. |
| [committing](.agents/skills/committing/) | Prepare conventional commit messages for this Go package. |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Evaluate and choose Go dependencies with alternatives and risk tradeoffs. |
| [go-best-practices](.agents/skills/go-best-practices/) | Apply Google Go style and architecture best practices to code changes. |
| [linting](.agents/skills/linting/) | Configure or run golangci-lint and fix lint issues in this package. |
| [modernizing](.agents/skills/modernizing/) | Adopt newer Go language and toolchain features safely. |
| [ralphy-initializing](.agents/skills/ralphy-initializing/) | Initialize or repair the .ralphy workflow configuration. |
| [ralphy-todo-creating](.agents/skills/ralphy-todo-creating/) | Generate or refine TODO tracking via the Ralphy workflow. |
| [readme-creating](.agents/skills/readme-creating/) | Create or rewrite README.md for this package. |
| [releasing](.agents/skills/releasing/) | Prepare release and semantic version workflows for this package. |
| [testing](.agents/skills/testing/) | Design or update tests (table-driven, fuzz, benchmark, and edge-case coverage). |

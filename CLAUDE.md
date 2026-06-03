# DeepClone

High-performance deep cloning library for Go values whose state can be represented as memory-owned data.

**Module**: `github.com/kaptinlin/deepclone`
**Go version**: see `go.mod`
**Dependencies**: stdlib only for core logic; `go-cmp` and `testify` for tests

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

Small public surface with the engine kept mostly in `clone.go`:

```text
clone.go              # Clone engine, fast paths, graph registry, struct metadata cache
cloner.go             # Strongly typed Cloner[T] protocol
errors.go             # UnsupportedError and stable path helpers
doc.go                # Package documentation
*_test.go             # Unit, edge, concurrent, cache, example, and benchmark tests
examples/             # Runnable examples
benchmarks/           # Separate comparison module
```

### Public API

```go
func Clone[T any](src T) (T, error)
func MustClone[T any](src T) T

type Cloner[T any] interface {
	Clone() (T, error)
}

type UnsupportedError struct {
	Path   string
	Type   reflect.Type
	Reason string
}
```

`CacheStats` and `ResetCache` are not public API. Cache tests use package-private `cacheStats` and `resetCache`.

## Package Contract

`Clone` either returns an independent clone or an error explaining why the value cannot be honestly cloned.

- Clone Go values whose state is ordinary memory-owned data.
- Preserve supported object relationships: pointer cycles, shared pointer targets, map identity, slice cycles, and supported field/array element pointer relations.
- Preserve nil semantics for nil pointers, slices, maps, interfaces, channels, functions, and unsafe pointers.
- Reject runtime/resource state instead of silently copying it.
- Do not use `unsafe` to read or write unexported fields.
- Let type authors own private invariants through `Cloner[T]`.

## Clone Flow

`Clone` checks paths in this order:

1. **Primitive fast path**: primitives return as-is with zero allocation.
2. **Scalar slice fast path**: common scalar slices use `cloneSliceExact[S, E]` with one allocation.
3. **Scalar map fast path**: simple maps use `maps.Clone`; `map[string]any` stays on the graph-aware path.
4. **Strong custom clone**: top-level values implementing `Cloner[T]` delegate to `Clone() (T, error)`.
5. **Reflection graph engine**: pointers, slices, maps, structs, arrays, and interfaces clone through a shared `cloneContext`.

Fast paths are allowed only when they preserve the same semantics as the reflection path.

## Custom Cloning

Implement `Cloner[T]` when a type owns private mutable state, resources, invariants, or a domain-specific cloning rule.

```go
func (d Document) Clone() (Document, error) {
	content, err := Clone(d.Content)
	if err != nil {
		return Document{}, err
	}
	return Document{Title: d.Title, Content: content}, nil
}
```

The reflection engine also recognizes concrete methods shaped like `Clone() (Concrete, error)` when cloning nested values. Circular reference detection does not apply inside custom clone methods; handle cycles there manually if needed.

Non-conforming `Clone` methods, such as `Clone() any`, are ignored by the custom clone protocol and cloned through normal reflection when possible.

## Struct Metadata Cache

- `structCache` maps `reflect.Type` to `structTypeInfo`.
- `structTypeInfo` stores per-field metadata: index, name, export status, and `copyField` or `cloneField` action.
- The cache is protected by `sync.RWMutex` with double-check locking.
- The cache is bounded by the number of distinct struct types seen.
- It is an implementation detail, not public observability state.

## Graph Engine

`cloneContext.visited` maps typed `visitKey` values to cloned `reflect.Value`s.

```go
type visitKey struct {
	kind visitKind
	addr uintptr
	typ  reflect.Type
}
```

Important rules:

- Register pointer targets before recursing to preserve cycles.
- Register cloned maps immediately after creation to preserve map cycles.
- Track slices only when element kind can contain cycles.
- Include type in slice and map visit keys to avoid address collisions.
- Register exported struct fields and array elements that can be addressed.
- Do not promise full backing-array alias reconstruction.
- Do not promise map entry interior pointer reconstruction.

## Unsupported State

Unsupported values return `UnsupportedError` with stable `Path`, `Type`, and `Reason`.

Examples:

- `$`
- `$.Config.Secrets[0]`
- `$["main"].Ch`
- `$.Workers["main"].Ch`

Rejected state:

- non-nil channels
- non-nil functions
- non-nil unsafe pointers
- sync primitives
- atomic runtime state
- file handles
- unexported reference-like fields

Nil channel/function/unsafe pointer values keep nil semantics and do not error.

## Unexported Fields

Struct cloning starts with a shallow copy, then recursively replaces safe exported fields.

- Unexported value-like fields are preserved by the shallow copy.
- Unexported reference-like fields return `UnsupportedError`.
- Resource/runtime fields return `UnsupportedError`.
- Types with private invariants should implement `Cloner[T]`.

## Testing

### Test Structure

- Table-driven tests with `t.Run()` subtests.
- Concurrent stress tests use 100-200 goroutines and hundreds of iterations.
- Cache tests use package-private `resetCache` for isolation.
- Edge tests cover committed object relationships.
- Tests should pass under `-race`.

### Test Files

```text
clone_test.go         # Core cloning, unsupported paths, Cloner, nils, cycles
edge_test.go          # Promised object relationship tests
cache_test.go         # Struct metadata cache behavior
concurrent_test.go    # Concurrent stress tests
benchmark_test.go     # Performance benchmarks
example_test.go       # Testable examples for GoDoc
```

### Promised Relationship Coverage

Keep tests for:

- primitive zero allocations and equality
- nil pointer/slice/map/interface/function/channel behavior
- empty slice/map distinct from nil
- shallow struct copy plus deep replacement of exported mutable fields
- private primitive preservation
- private reference-like rejection
- pointer cycles
- slice/interface cycles
- map cycles
- typed nil interface values
- pointer to struct field
- pointer to array element
- map key/value sharing the same pointer object
- repeated fields sharing one pointer/map/slice object
- `Cloner[T]` success and error propagation
- non-conforming `Clone` methods ignored by custom clone protocol
- channel/function/unsafe pointer/sync/file rejection
- concurrent clone and metadata cache race safety

Do not add tests that turn distinct subslice backing-array aliasing or map entry interior pointers into public contract.

## Performance

### Optimization Principles

- **Primitive values**: 0 allocs.
- **Scalar slices**: 1 allocation.
- **Scalar maps**: stay close to `maps.Clone`.
- **Structs**: cache metadata after first analysis.
- **Graph clone**: overhead should scale with object count.

Recent sanity benchmark on darwin/arm64:

| Operation | Performance | Memory | Allocations |
|-----------|-------------|--------|-------------|
| int | 1.7 ns/op | 0 B/op | 0 allocs/op |
| string | 1.9 ns/op | 0 B/op | 0 allocs/op |
| slice 100 | 160 ns/op | 896 B/op | 1 alloc/op |
| map 100 | 390 ns/op | 3,544 B/op | 4 allocs/op |
| large slice 10K | 4,988 ns/op | 81,921 B/op | 1 alloc/op |

Run `make bench-comparison` for detailed comparisons with other libraries.

## Coding Rules

### Features Available Under the Declared Go Version

- Use `for range N` instead of `for i := 0; i < N; i++`.
- Use `b.Loop()` in benchmarks instead of `for i := 0; i < b.N; i++`.
- Use `reflect.MakeMapWithSize(type, size)` for map initialization.
- Use `clear(map)` for map clearing.

### Conventions

- All tests must pass with `-race`.
- Use `any` instead of `interface{}`.
- Use `reflect.Pointer`, not deprecated `reflect.Ptr`.
- Test assertions use `testify/assert` and `testify/require`.
- Keep core runtime dependencies stdlib-only.

## Error Handling

This library does not silently best-effort clone unsupported state.

- **Invalid values**: return as-is.
- **Nil values**: return as-is.
- **Unsupported values**: return `UnsupportedError`.
- **Custom clone errors**: propagate unchanged.
- **Type conversion failures**: return `UnsupportedError`.

Error strings should be stable, short, and avoid dumping values or secrets.

## Linting

- **Version**: Managed via `.golangci.version` (currently 2.12.2)
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

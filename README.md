# DeepClone

A high-performance deep cloning library for Go values whose state can be represented as memory-owned data.

[![Go Module](https://img.shields.io/badge/go-1.26.4%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/deepclone)](https://goreportcard.com/report/github.com/kaptinlin/deepclone)

DeepClone is intentionally honest: it clones supported in-memory values, preserves supported object relationships, and returns an error for runtime or resource state that cannot be meaningfully deep-cloned.

## Features

- **Generic API**: Clone Go values with `deepclone.Clone(value)`.
- **Honest errors**: Unsupported state returns `UnsupportedError` with path, type, and reason.
- **Fast common paths**: Primitives, scalar slices, and scalar maps avoid reflection overhead.
- **Graph safety**: Pointer cycles, map cycles, slice cycles, and shared pointer targets are preserved.
- **Custom cloning**: Implement `Cloner[T]` for private invariants or domain-specific behavior.
- **Concurrent use**: Package state is safe under concurrent cloning.
- **No runtime dependencies**: Core library uses only the Go standard library.

## Installation

```bash
go get github.com/kaptinlin/deepclone
```

Requires Go 1.26.4 or later.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

func main() {
	original := map[string][]int{
		"numbers": {1, 2, 3},
		"scores":  {85, 90, 95},
	}

	cloned, err := deepclone.Clone(original)
	if err != nil {
		panic(err)
	}

	original["numbers"][0] = 999

	fmt.Println(original["numbers"])
	fmt.Println(cloned["numbers"])
}
```

Output:

```text
[999 2 3]
[1 2 3]
```

## API

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

Use `Clone` in production paths where unsupported state should be handled. Use `MustClone` for tests, fixtures, and values that are already known to be supported.

## Usage

### Clone structs and collections

```go
type User struct {
	Name    string
	Friends []string
	Config  map[string]any
}

user := User{
	Name:    "Alice",
	Friends: []string{"Bob", "Charlie"},
	Config:  map[string]any{"theme": "dark"},
}

cloned, err := deepclone.Clone(user)
if err != nil {
	return err
}

cloned.Friends[0] = "Eve"
```

`user.Friends` remains unchanged because supported slices, maps, arrays, structs, pointers, and interfaces are cloned recursively.

### Handle unsupported state

```go
type Worker struct {
	Ch chan int
}

_, err := deepclone.Clone(map[string]Worker{
	"main": {Ch: make(chan int)},
})
if err != nil {
	var unsupported *deepclone.UnsupportedError
	if errors.As(err, &unsupported) {
		fmt.Println(unsupported.Path)   // $["main"].Ch
		fmt.Println(unsupported.Reason) // channels cannot be cloned
	}
}
```

Unsupported errors are stable and intentionally do not include value contents.

### Customize clone behavior

```go
type Document struct {
	Title   string
	Content []byte
	Version int
}

func (d Document) Clone() (Document, error) {
	content, err := deepclone.Clone(d.Content)
	if err != nil {
		return Document{}, err
	}
	return Document{
		Title:   d.Title,
		Content: content,
		Version: d.Version + 1,
	}, nil
}
```

Types that implement `Cloner[T]` control their own cloning behavior. Circular reference detection does not apply inside custom `Clone` methods.

## Semantics

DeepClone preserves supported object relationships:

- pointer cycles
- map cycles
- slice cycles
- shared pointer targets
- pointer-to-struct-field relationships when the owner is cloned in the same graph
- pointer-to-array-element relationships when the owner is cloned in the same graph

DeepClone does not promise full backing-array alias reconstruction for distinct subslices, and it does not promise map entry interior pointer reconstruction.

## Special Cases

| Value kind | Clone behavior |
| --- | --- |
| Nil pointers, slices, maps, interfaces, channels, functions, unsafe pointers | Preserved as nil |
| Non-nil channels | Return `UnsupportedError` |
| Non-nil functions | Return `UnsupportedError` |
| Non-nil unsafe pointers | Return `UnsupportedError` |
| Sync primitives and atomic state | Return `UnsupportedError` |
| File handles | Return `UnsupportedError` |
| Unexported value-like struct fields | Preserved by shallow struct copy |
| Unexported reference-like struct fields | Return `UnsupportedError`; implement `Cloner[T]` for private invariants |

## Performance

DeepClone keeps common operations fast with primitive, scalar slice, and scalar map fast paths plus cached reflection metadata for structs.

Recent sanity benchmark on darwin/arm64:

| Operation | Performance | Memory | Allocations |
| --- | ---: | ---: | ---: |
| int | 1.7 ns/op | 0 B/op | 0 allocs/op |
| string | 1.9 ns/op | 0 B/op | 0 allocs/op |
| slice 100 | 160 ns/op | 896 B/op | 1 alloc/op |
| map 100 | 390 ns/op | 3,544 B/op | 4 allocs/op |
| large slice 10K | 4,988 ns/op | 81,921 B/op | 1 alloc/op |

```bash
task bench
```

Run comparison benchmarks against other clone libraries:

```bash
task bench-comparison
```

## Development

```bash
task deps   # Download and tidy dependencies
task lint   # Run golangci-lint and tidy checks
task test   # Run tests with -race
task verify # Run the full local verification pipeline
```

## Examples

See [examples/](examples/) for runnable examples covering basic values, circular references, and custom clone behavior.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

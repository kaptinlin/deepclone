# DeepClone

A high-performance deep cloning library for Go that safely copies arbitrary values with one generic API

[![Go Module](https://img.shields.io/badge/go-1.26.3%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/deepclone)](https://goreportcard.com/report/github.com/kaptinlin/deepclone)

## Features

- **Generic API**: Clone any Go value with `deepclone.Clone(value)`
- **Fast common paths**: Copy primitives, common slices, and common maps without reflection overhead
- **Circular reference safety**: Preserve object graphs when cloning through the reflection path
- **Custom cloning**: Implement `Cloneable` for domain-specific copy behavior
- **Concurrent use**: Share the package safely across goroutines
- **No runtime dependencies**: Core library uses only the Go standard library

## Installation

```bash
go get github.com/kaptinlin/deepclone
```

Requires Go 1.26.3 or later.

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

    cloned := deepclone.Clone(original)
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

## API Overview

| API | Purpose |
| --- | --- |
| `Clone[T any](src T) T` | Return a deep copy of `src` |
| `Cloneable` | Let a type provide its own clone implementation |
| `CacheStats() (entries, fields int)` | Inspect cached struct metadata |
| `ResetCache()` | Clear cached struct metadata, usually in tests and benchmarks |

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

cloned := deepclone.Clone(user)
cloned.Friends[0] = "Eve"
```

`user.Friends` remains unchanged because slices, maps, arrays, structs, pointers, and interfaces are cloned recursively.

### Customize clone behavior

```go
type Document struct {
    Title   string
    Content []byte
    Version int
}

func (d Document) Clone() any {
    return Document{
        Title:   d.Title,
        Content: deepclone.Clone(d.Content),
        Version: d.Version + 1,
    }
}
```

Types that implement `Cloneable` control their own cloning behavior. Circular reference detection does not apply inside custom `Clone` methods.

### Special cases

| Value kind | Clone behavior |
| --- | --- |
| Nil pointers, slices, and maps | Preserved as nil |
| Functions | Returned as-is |
| Channels | Returned as the zero value of the same channel type |
| Unexported struct fields | Left at the zero value |

## Performance

DeepClone keeps common operations fast with primitive, slice, and map fast paths plus cached reflection metadata for structs.

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

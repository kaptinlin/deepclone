# DeepClone

A high-performance deep cloning library for Go that provides safe, efficient copying of any Go value.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/deepclone)](https://goreportcard.com/report/github.com/kaptinlin/deepclone)

## Features

- **High Performance**: Zero-allocation fast paths for primitive types
- **Circular Reference Safe**: Automatic detection and handling of circular references
- **Thread Safe**: Concurrent operations with safe caching mechanisms
- **Universal Support**: Works with all Go types including channels, functions, and interfaces
- **Extensible**: Custom cloning behavior via `Cloneable` interface
- **Zero Dependencies**: Uses only Go standard library

## Installation

```bash
go get github.com/kaptinlin/deepclone
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/deepclone"
)

func main() {
    // Deep clone any value
    original := map[string][]int{
        "numbers": {1, 2, 3},
        "scores":  {85, 90, 95},
    }

    cloned := deepclone.Clone(original)

    // Modify original - cloned remains independent
    original["numbers"][0] = 999

    fmt.Println("Original:", original["numbers"]) // [999, 2, 3]
    fmt.Println("Cloned:", cloned["numbers"])     // [1, 2, 3]
}
```

## Core Concept

All operations perform **deep copies** by default:

- **Primitives**: `int`, `string`, `bool` -- Copied by value (zero allocations)
- **Collections**: `slice`, `map`, `array` -- New containers with cloned elements
- **Structs**: New instances with all fields deeply cloned
- **Pointers**: New pointers pointing to cloned values
- **Custom Types**: Support via `Cloneable` interface

## Examples

### Basic Usage

```go
// Primitives (zero allocation)
number := deepclone.Clone(42)
text := deepclone.Clone("hello")

// Collections (deep cloned)
slice := deepclone.Clone([]string{"a", "b", "c"})
data := deepclone.Clone(map[string]int{"key": 42})

// Complex structures
type User struct {
    Name    string
    Friends []string
    Config  map[string]interface{}
}

user := User{
    Name:    "Alice",
    Friends: []string{"Bob", "Charlie"},
    Config:  map[string]interface{}{"theme": "dark"},
}

cloned := deepclone.Clone(user) // Complete deep copy
```

### Custom Cloning Behavior

```go
type Document struct {
    Title   string
    Content []byte
    Version int
}

// Implement custom cloning logic
func (d Document) Clone() any {
    return Document{
        Title:   d.Title,
        Content: deepclone.Clone(d.Content).([]byte),
        Version: d.Version + 1, // Increment version on clone
    }
}

doc := Document{Title: "My Doc", Version: 1}
cloned := deepclone.Clone(doc) // Version becomes 2
```

For more examples, see **[examples/](examples/)** directory.

## Performance

DeepClone is optimized for performance with:

- **Zero allocations** for primitive types
- **Fast paths** for common slice/map types
- **Reflection caching** for struct types
- **Minimal overhead** for complex operations

### Benchmark Results

Tested on Apple M3, macOS (darwin/arm64):

| Operation | Performance | Memory | Allocations |
|-----------|-------------|---------|-------------|
| Primitives (int/string/bool) | 2.7-3.6 ns/op | 0 B/op | 0 allocs/op |
| Slice (100 ints) | 200.6 ns/op | 896 B/op | 1 allocs/op |
| Map (100 entries) | 4,299 ns/op | 3,544 B/op | 4 allocs/op |
| Simple Struct | 248.6 ns/op | 128 B/op | 4 allocs/op |
| Nested Struct | 1,386 ns/op | 952 B/op | 19 allocs/op |
| Large Slice (10K ints) | 6,709 ns/op | 81,920 B/op | 1 allocs/op |

For detailed benchmarks and comparisons with other libraries, see **[benchmarks/](benchmarks/)**.

```bash
# Run benchmarks
cd benchmarks && go test -bench=. -benchmem
```

## API Reference

### Core Function

```go
func Clone[T any](src T) T
```

Creates a deep copy of any value. The returned value is completely independent of the original.

### Custom Cloning Interface

```go
type Cloneable interface {
    Clone() any
}
```

Implement this interface to provide custom cloning behavior for your types.

## Advanced Features

- **Circular Reference Detection**: Prevents infinite loops in self-referencing structures
- **Interface Preservation**: Maintains original interface types while cloning concrete values
- **Thread Safety**: All operations are safe for concurrent use
- **Type Caching**: Struct metadata is cached for improved performance on repeated operations

## Contributing

We welcome contributions! Please feel free to:

- Report bugs
- Suggest new features
- Submit pull requests
- Improve documentation

## Requirements

- Go 1.25 or later
- No external dependencies

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

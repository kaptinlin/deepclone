# Go Deep Clone Performance Benchmarks

Performance comparison for Go deep clone libraries.

## Libraries

| Library | Package |
|---------|---------|
| **this** | `github.com/kaptinlin/deepclone` |
| **mohae** | `github.com/mohae/deepcopy` |
| **jinzhu** | `github.com/jinzhu/copier` |
| **huandu** | `github.com/huandu/go-clone` |
| **golang-design** | `golang.design/x/reflect` |

## Environment

- **Platform**: Apple M3, macOS (darwin/arm64)
- **Test Method**: `go test -bench=. -benchmem -benchtime=1s`

## Results

### Basic Types

| Type | Library | ns/op | B/op | allocs/op |
|------|---------|-------|------|-----------|
| **Int** | | | | |
| | this | 1.47 | 0 | 0 |
| | mohae | 37.29 | 16 | 2 |
| | huandu | 10.10 | 0 | 0 |
| | golang-design | 19.93 | 32 | 1 |
| **String** | | | | |
| | this | 1.55 | 0 | 0 |
| | mohae | 54.56 | 48 | 3 |
| | huandu | 18.10 | 16 | 1 |
| | golang-design | 51.47 | 80 | 4 |
| **Float64** | | | | |
| | this | 1.46 | 0 | 0 |
| | mohae | 42.61 | 24 | 3 |
| | huandu | 13.85 | 8 | 1 |
| | golang-design | 26.04 | 40 | 2 |
| **Bool** | | | | |
| | this | 1.62 | 0 | 0 |
| | mohae | 34.44 | 2 | 2 |
| | huandu | 6.96 | 0 | 0 |
| | golang-design | 19.12 | 32 | 1 |

### Collections

| Collection | Library | ns/op | B/op | allocs/op |
|------------|---------|-------|------|-----------|
| **Slice (100 ints)** | | | | |
| | this | 69.41 | 896 | 1 |
| | mohae | 2334 | 1792 | 105 |
| | huandu | 118.1 | 944 | 3 |
| | golang-design | 2381 | 1776 | 104 |
| **Map (100 entries)** | | | | |
| | this | 1636 | 3544 | 4 |
| | mohae | 14772 | 16048 | 513 |
| | huandu | 5961 | 5944 | 204 |
| | golang-design | 8129 | 8376 | 405 |

### Complex Structures

| Structure | Library | ns/op | B/op | allocs/op |
|-----------|---------|-------|------|-----------|
| **Simple Struct** | | | | |
| | this | 194 | 160 | 5 |
| | mohae | 138 | 96 | 3 |
| | jinzhu | 908 | 456 | 15 |
| | huandu | 85.5 | 96 | 3 |
| | golang-design | 160 | 148 | 6 |
| **Nested Struct** | | | | |
| | this | 764 | 952 | 19 |
| | mohae | 1144 | 1176 | 35 |
| | jinzhu | 4287 | 2336 | 68 |
| | huandu | 693 | 776 | 16 |
| | golang-design | 1259 | 1280 | 45 |

### Special Cases

| Case | Library | ns/op | B/op | allocs/op |
|------|---------|-------|------|-----------|
| **Pointers** | | | | |
| | this | 117 | 64 | 2 |
| | mohae | 210 | 136 | 7 |
| | huandu | 71.6 | 32 | 1 |
| | golang-design | 218 | 180 | 7 |
| **Circular Reference** | | | | |
| | this | 122 | 64 | 2 |
| | golang-design | 216 | 184 | 7 |
| **Interface** | | | | |
| | this | 157 | 64 | 2 |
| | mohae | 142 | 64 | 2 |
| | huandu | 72.2 | 64 | 2 |
| | golang-design | 146 | 116 | 5 |

### Large Data Performance

| Test | Library | ns/op | B/op | allocs/op |
|------|---------|-------|------|-----------|
| **Large Slice (10K ints)** | | | | |
| | this | 3,654 | 81,920 | 1 |
| | mohae | 221,200 | 162,017 | 10,005 |
| | huandu | 6,079 | 81,968 | 3 |
| | golang-design | 229,766 | 162,001 | 10,004 |

## Performance Summary

### Performance Characteristics by Library

| Library | Strengths | Use Cases |
|---------|-----------|-----------|
| **this** | Zero-allocation primitives, efficient large data handling | High-performance applications requiring minimal allocations |
| **mohae** | Simple API, broad compatibility | General-purpose copying where performance is not critical |
| **jinzhu** | Field mapping, cross-type copying, struct tags support | Data transformation and mapping between different struct types |
| **huandu** | Excellent struct performance, minimal allocations | High-frequency struct cloning with good performance requirements |
| **golang-design** | Experimental features, comprehensive type support | Research and experimental applications |

## Usage

```bash
cd benchmarks
go test -bench=. -benchmem -benchtime=1s
```

### Test Scenarios

- **Primitives**: Basic Go types (int, string, bool, float64)
- **Collections**: Slices, maps, arrays of various sizes
- **Structures**: Simple and nested struct copying
- **Pointers**: Single and multi-level pointer dereferencing
- **Circular References**: Self-referencing data structures
- **Interface Types**: Copying through interfaces
- **Large Data**: Performance with larger datasets 
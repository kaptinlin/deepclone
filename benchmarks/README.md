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
| | this | 2.98 | 0 | 0 |
| | mohae | 65.51 | 16 | 2 |
| | huandu | 12.29 | 0 | 0 |
| | golang-design | 32.76 | 32 | 1 |
| **String** | | | | |
| | this | 2.66 | 0 | 0 |
| | mohae | 95.27 | 48 | 3 |
| | huandu | 31.87 | 16 | 1 |
| | golang-design | 92.90 | 80 | 4 |
| **Float64** | | | | |
| | this | 2.72 | 0 | 0 |
| | mohae | 77.85 | 24 | 3 |
| | huandu | 24.39 | 8 | 1 |
| | golang-design | 60.31 | 40 | 2 |
| **Bool** | | | | |
| | this | 3.62 | 0 | 0 |
| | mohae | 78.05 | 2 | 2 |
| | huandu | 15.67 | 0 | 0 |
| | golang-design | 52.80 | 32 | 1 |

### Collections

| Collection | Library | ns/op | B/op | allocs/op |
|------------|---------|-------|------|-----------|
| **Slice (100 ints)** | | | | |
| | this | 200.6 | 896 | 1 |
| | mohae | 5,301 | 1,792 | 105 |
| | huandu | 330.2 | 944 | 3 |
| | golang-design | 4,840 | 1,776 | 104 |
| **Map (100 entries)** | | | | |
| | this | 4,299 | 3,544 | 4 |
| | mohae | 39,568 | 16,048 | 513 |
| | huandu | 12,212 | 5,944 | 204 |
| | golang-design | 21,517 | 8,376 | 405 |

### Complex Structures

| Structure | Library | ns/op | B/op | allocs/op |
|-----------|---------|-------|------|-----------|
| **Simple Struct** | | | | |
| | this | 248.6 | 128 | 4 |
| | mohae | 227.3 | 96 | 3 |
| | jinzhu | 1,642 | 456 | 15 |
| | huandu | 178.5 | 96 | 3 |
| | golang-design | 325.4 | 148 | 6 |
| **Nested Struct** | | | | |
| | this | 1,386 | 952 | 19 |
| | mohae | 2,143 | 1,176 | 35 |
| | jinzhu | 6,660 | 2,336 | 68 |
| | huandu | 1,059 | 776 | 16 |
| | golang-design | 2,465 | 1,280 | 45 |

### Special Cases

| Case | Library | ns/op | B/op | allocs/op |
|------|---------|-------|------|-----------|
| **Pointers** | | | | |
| | this | 200.8 | 64 | 2 |
| | mohae | 343.4 | 136 | 7 |
| | huandu | 181.3 | 32 | 1 |
| | golang-design | 446.1 | 180 | 7 |
| **Circular Reference** | | | | |
| | this | 212.3 | 64 | 2 |
| | golang-design | 411.1 | 184 | 7 |
| **Interface** | | | | |
| | this | 190.9 | 64 | 2 |
| | mohae | 200.1 | 64 | 2 |
| | huandu | 135.5 | 64 | 2 |
| | golang-design | 347.8 | 116 | 5 |

### Large Data Performance

| Test | Library | ns/op | B/op | allocs/op |
|------|---------|-------|------|-----------|
| **Large Slice (10K ints)** | | | | |
| | this | 6,709 | 81,920 | 1 |
| | mohae | 404,820 | 162,017 | 10,005 |
| | huandu | 12,002 | 81,968 | 3 |
| | golang-design | 395,634 | 162,001 | 10,004 |

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
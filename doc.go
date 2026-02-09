// Package deepclone provides high-performance deep cloning functionality for Go.
//
// This package uses Go generics, reflection caching, and careful memory
// management to achieve zero-allocation hot paths where possible.
//
// Basic Usage:
//
//	// Clone any value with full deep copy semantics
//	original := []int{1, 2, 3}
//	cloned := deepclone.Clone(original)
//
//	// Clone structs with nested data
//	type Config struct {
//	    Tags []string
//	    Meta map[string]int
//	}
//	cfg := Config{Tags: []string{"a"}, Meta: map[string]int{"x": 1}}
//	copy := deepclone.Clone(cfg)
//
// Custom Cloning:
//
// Types that implement the [Cloneable] interface receive custom cloning
// behavior instead of the default reflection-based deep clone:
//
//	func (m MyStruct) Clone() any {
//	    return MyStruct{
//	        Name: m.Name,
//	        Data: deepclone.Clone(m.Data),
//	    }
//	}
//
// Performance Hierarchy:
//
// Clone uses a hierarchical optimization strategy:
//   - Ultra-fast path: primitive types (zero allocation, direct return)
//   - Fast path: common slice and map types (generic copy, no reflection)
//   - Cloneable path: types implementing [Cloneable] interface
//   - Reflection path: all other types (cached struct info, circular ref detection)
//
// Supported Types:
//   - All primitive types (int, string, bool, float, complex, etc.)
//   - Slices, maps, and arrays (with deep cloning of elements)
//   - Pointers and pointer chains (with circular reference detection)
//   - Structs (with automatic field-by-field deep cloning)
//   - Interfaces (with concrete type preservation)
//   - Custom types implementing [Cloneable] interface
//   - Channels return zero value; functions return as-is
//
// Thread Safety:
//
// All cloning operations are safe for concurrent use. The internal struct
// type cache uses [sync.RWMutex] with double-check locking for thread-safe
// access without contention on the hot path.
package deepclone

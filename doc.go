// Package deepclone provides high-performance deep cloning functionality for Go.
//
// This package is designed for extreme performance optimization, utilizing Go 1.21+
// generics, reflection caching, and careful memory management to achieve zero-allocation
// hot paths where possible.
//
// Basic Usage:
//
//	import "github.com/kaptinlin/deepclone"
//
//	// Clone any value with deep copying semantics
//	dst := deepclone.Clone(src)
//
//	// For custom types, implement the Cloneable interface for deep cloning
//	type MyStruct struct {
//	    Name string
//	    Data []int
//	}
//
//	func (m MyStruct) Clone() any {
//	    return MyStruct{
//	        Name: m.Name,
//	        Data: deepclone.Clone(m.Data).([]int), // Deep clone nested data
//	    }
//	}
//
// Performance Features:
//   - Zero-allocation hot paths for primitive types
//   - Reflection result caching for struct types
//   - Optimized fast paths for common slice and map types
//   - CPU cache-friendly data access patterns
//
// Supported Types:
//   - All primitive types (int, string, bool, etc.)
//   - Slices, maps, and arrays (with deep cloning of elements)
//   - Pointers and pointer chains (with circular reference detection)
//   - Structs (with automatic field-by-field deep cloning)
//   - Interfaces (with concrete type preservation)
//   - Custom types implementing Cloneable interface
//
// Deep Cloning Semantics:
//   - All cloning operations perform deep copies by default
//   - Nested data structures are recursively cloned
//   - Circular references are safely detected and handled
//   - Custom types can override default behavior via Cloneable interface
//
// Thread Safety:
//   - All cloning operations are thread-safe
//   - Internal caches use concurrent-safe mechanisms
//   - No global state modifications during cloning
package deepclone

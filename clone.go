package deepclone

import (
	"maps"
	"reflect"
	"sync"
)

// A fieldAction indicates whether a struct field needs deep cloning or simple copy.
type fieldAction int

const (
	copyField  fieldAction = iota // Simple assignment (primitive types)
	cloneField                    // Needs deep cloning (complex types)
)

var (
	// structCache caches struct type information to avoid repeated reflection.
	//
	// Memory analysis: LRU eviction is unnecessary because the cache is keyed
	// by reflect.Type, which is interned by the Go runtime — each distinct
	// struct type has exactly one reflect.Type value. The number of entries is
	// bounded by the number of distinct struct types the program clones, which
	// is finite and determined at compile time. Even a large application with
	// 1000 struct types averaging 10 fields uses ~1.3 MB — negligible for any
	// program that uses reflection. Adding LRU would hurt hot-path performance
	// (linked-list pointer updates + extra locking) for no practical benefit.
	//
	// Use ResetCache to reclaim memory if needed (e.g., in tests).
	structCache = make(map[reflect.Type]*structTypeInfo)
	cacheMutex  sync.RWMutex
)

// A cloneContext tracks visited objects to prevent infinite loops in circular references.
type cloneContext struct {
	visited map[uintptr]reflect.Value
}

func newCloneContext() *cloneContext {
	return &cloneContext{
		visited: make(map[uintptr]reflect.Value, 8), // Pre-allocate for common cases
	}
}

// A structTypeInfo caches per-field clone decisions for a struct type
// so that repeated cloning of the same type avoids redundant reflection.
type structTypeInfo struct {
	actions []fieldAction
	fields  []reflect.StructField
}

// getStructTypeInfo returns cached struct field information for the given type,
// computing and caching it on first access.
func getStructTypeInfo(t reflect.Type) *structTypeInfo {
	cacheMutex.RLock()
	if info, exists := structCache[t]; exists {
		cacheMutex.RUnlock()
		return info
	}
	cacheMutex.RUnlock()

	// Compute field actions
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check in case another goroutine computed it
	if info, exists := structCache[t]; exists {
		return info
	}

	numFields := t.NumField()
	actions := make([]fieldAction, numFields)
	fields := make([]reflect.StructField, numFields)

	for i := range numFields {
		field := t.Field(i)
		fields[i] = field

		if !field.IsExported() {
			actions[i] = copyField // Skip unexported fields
			continue
		}

		// Determine if field needs deep cloning or simple copy.
		switch field.Type.Kind() {
		case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Array, reflect.Struct:
			actions[i] = cloneField
		case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Chan, reflect.Func, reflect.String, reflect.UnsafePointer:
			actions[i] = copyField
		}
	}

	info := &structTypeInfo{
		actions: actions,
		fields:  fields,
	}
	structCache[t] = info
	return info
}

// CacheStats returns the number of struct types currently cached (entries)
// and the total number of cached fields across all types (fields).
//
// This is useful for monitoring cache growth and validating that
// memory usage remains bounded. In practice, the entry count equals
// the number of distinct struct types cloned by the program, which
// is finite and compile-time determined.
//
// CacheStats is safe for concurrent use.
func CacheStats() (entries, fields int) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entries = len(structCache)
	for _, info := range structCache {
		fields += len(info.fields)
	}
	return entries, fields
}

// ResetCache clears the struct type cache, releasing all cached
// reflection data. Subsequent [Clone] operations will re-populate
// the cache on demand.
//
// This is primarily useful in tests or long-running applications
// that dynamically load and unload plugins with unique struct types.
// Under normal usage the cache is small (bounded by the number of
// distinct struct types) and does not need to be reset.
//
// ResetCache is safe for concurrent use.
func ResetCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	clear(structCache)
}

// cloneSliceExact creates a copy of the slice with exact capacity preservation.
// The compiler inlines this generic helper for performance.
func cloneSliceExact[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	cloned := make(S, len(s), cap(s))
	copy(cloned, s)
	return cloned
}

// Clone creates a deep copy of the given value, preserving the complete
// object graph including circular references.
//
// Clone uses a hierarchical optimization strategy for maximum performance:
//   - Primitive types (int, string, bool, etc.) return as-is with zero allocation
//   - Common slice types ([]int, []string, []byte, etc.) use optimized generic copy
//   - Common map types (map[string]string, etc.) use [maps.Clone]
//   - Types implementing [Cloneable] delegate to their Clone method
//   - All other types use cached reflection with circular reference detection
//
// Special cases:
//   - nil values return as-is
//   - Channels return the zero value of their type
//   - Functions return as-is (cannot be deep cloned)
//   - Slice capacity and map size hints are preserved
//
// Usage:
//
//	dst := deepclone.Clone(src)
func Clone[T any](src T) T {
	// Ultra-fast paths for primitive types - avoid all reflection and allocations
	switch any(src).(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128:
		// Primitive types are value types, return as-is (zero allocations)
		return src
	case string:
		// Strings are immutable in Go, return as-is (zero allocations)
		return src
	}

	// Fast paths for common slice types using generic cloneSliceExact helper.
	// We use this to ensure exact capacity preservation which is required by strict tests.
	switch s := any(src).(type) {
	case []int:
		return any(cloneSliceExact(s)).(T)
	case []int8:
		return any(cloneSliceExact(s)).(T)
	case []int16:
		return any(cloneSliceExact(s)).(T)
	case []int32:
		return any(cloneSliceExact(s)).(T)
	case []int64:
		return any(cloneSliceExact(s)).(T)
	case []uint:
		return any(cloneSliceExact(s)).(T)
	case []uint16:
		return any(cloneSliceExact(s)).(T)
	case []uint32:
		return any(cloneSliceExact(s)).(T)
	case []uint64:
		return any(cloneSliceExact(s)).(T)
	case []float32:
		return any(cloneSliceExact(s)).(T)
	case []float64:
		return any(cloneSliceExact(s)).(T)
	case []string:
		return any(cloneSliceExact(s)).(T)
	case []bool:
		return any(cloneSliceExact(s)).(T)
	case []byte:
		return any(cloneSliceExact(s)).(T)
	}

	// Fast paths for simple map types using generic maps package (Go 1.21+).
	// map[string]any is deliberately excluded to handle potential circular references via reflection.
	switch m := any(src).(type) {
	case map[string]int:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[string]string:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[string]float64:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[string]bool:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[int]int:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[int]string:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	case map[int]bool:
		if m == nil {
			return src
		}
		return any(maps.Clone(m)).(T)
	}

	// Reflection-based path for complex types.
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src
	}

	// Check if type implements Cloneable interface before reflection-based cloning.
	if cloneable, ok := any(src).(Cloneable); ok {
		if result, ok := cloneable.Clone().(T); ok {
			return result
		}
	}

	// Fast path for nil pointers without additional reflection.
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return src
	}

	// Use reflection-based cloning for complex types with circular reference detection
	cc := newCloneContext()
	cloned := cc.cloneValue(v)
	if cloned.IsValid() {
		return cloned.Interface().(T)
	}

	return src
}

// cloneValue performs deep cloning of v using reflection.
func (cc *cloneContext) cloneValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}

	switch v.Kind() {
	case reflect.Pointer:
		return cc.clonePointer(v)

	case reflect.Slice:
		return cc.cloneSlice(v)

	case reflect.Map:
		return cc.cloneMap(v)

	case reflect.Struct:
		return cc.cloneStruct(v)

	case reflect.Array:
		return cc.cloneArray(v)

	case reflect.Interface:
		return cc.cloneInterface(v)

	case reflect.Chan:
		// Channels cannot be meaningfully cloned; return nil channel of same type.
		return reflect.Zero(v.Type())

	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Func, reflect.String, reflect.UnsafePointer:
		return v
	}

	// Unreachable: all reflect.Kind values are handled above.
	return v
}

// clonePointer creates a deep copy of a pointer, recursively cloning the pointed value.
func (cc *cloneContext) clonePointer(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Check for circular reference.
	addr := v.Pointer()
	if cloned, exists := cc.visited[addr]; exists {
		return cloned
	}

	// Create new pointer and clone the pointed value.
	elemType := v.Type().Elem()
	newPtr := reflect.New(elemType)

	// Store the new pointer in visited map before cloning the element
	// to handle self-referencing structures.
	cc.visited[addr] = newPtr

	clonedElem := cc.cloneValue(v.Elem())
	if clonedElem.IsValid() {
		newPtr.Elem().Set(clonedElem)
	}

	return newPtr
}

// cloneSlice creates a deep copy of a slice.
func (cc *cloneContext) cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Only track slices that might contain cycles or references.
	// This avoids map overhead for primitive slices (e.g., []int, []byte).
	elemKind := v.Type().Elem().Kind()
	needsTracking := elemKind == reflect.Pointer || elemKind == reflect.Interface ||
		elemKind == reflect.Slice || elemKind == reflect.Map || elemKind == reflect.Struct

	if needsTracking {
		addr := v.Pointer()
		if cloned, exists := cc.visited[addr]; exists {
			// Only return cached slice if it has the same length/capacity.
			// This prevents aliasing bugs where a sub-slice gets replaced by the full slice.
			if cloned.Len() == v.Len() && cloned.Cap() == v.Cap() {
				return cloned
			}
		}
	}

	length := v.Len()
	capacity := v.Cap()
	newSlice := reflect.MakeSlice(v.Type(), length, capacity)

	if needsTracking {
		cc.visited[v.Pointer()] = newSlice
	}

	for i := range length {
		clonedElem := cc.cloneValue(v.Index(i))
		if clonedElem.IsValid() {
			newSlice.Index(i).Set(clonedElem)
		}
	}

	return newSlice
}

// cloneMap creates a deep copy of a map.
func (cc *cloneContext) cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Check for circular reference.
	addr := v.Pointer()
	if cloned, exists := cc.visited[addr]; exists {
		return cloned
	}

	// Create new map with size hint for Swiss Tables optimization.
	size := v.Len()
	newMap := reflect.MakeMapWithSize(v.Type(), size)

	// Store in visited map for circular reference detection.
	cc.visited[addr] = newMap

	// Copy all key-value pairs with deep cloning.
	elemType := v.Type().Elem()
	for _, key := range v.MapKeys() {
		clonedKey := cc.cloneValue(key)
		clonedValue := cc.cloneValue(v.MapIndex(key))

		if !clonedKey.IsValid() || !clonedValue.IsValid() {
			continue
		}

		// Handle type alias conversions to prevent panic during SetMapIndex.
		if clonedValue.Type() != elemType {
			switch {
			case clonedValue.Type().ConvertibleTo(elemType):
				clonedValue = clonedValue.Convert(elemType)
			case clonedValue.Type().AssignableTo(elemType):
				// Assignable but not convertible; allow as-is.
			default:
				// Incompatible types; skip entry to avoid panic.
				continue
			}
		}
		newMap.SetMapIndex(clonedKey, clonedValue)
	}

	return newMap
}

// cloneStruct creates a deep copy of a struct using cached type information.
func (cc *cloneContext) cloneStruct(v reflect.Value) reflect.Value {
	structType := v.Type()
	newStruct := reflect.New(structType).Elem()

	info := getStructTypeInfo(structType)

	for i, action := range info.actions {
		if !info.fields[i].IsExported() {
			continue
		}

		srcField := v.Field(i)
		dstField := newStruct.Field(i)

		switch action {
		case copyField:
			if dstField.CanSet() {
				dstField.Set(srcField)
			}
		case cloneField:
			clonedField := cc.cloneValue(srcField)
			if clonedField.IsValid() && dstField.CanSet() {
				// Handle type alias conversions to prevent panic.
				if clonedField.Type() != dstField.Type() && clonedField.Type().ConvertibleTo(dstField.Type()) {
					clonedField = clonedField.Convert(dstField.Type())
				}
				dstField.Set(clonedField)
			}
		}
	}

	return newStruct
}

// cloneArray creates a deep copy of an array.
func (cc *cloneContext) cloneArray(v reflect.Value) reflect.Value {
	newArray := reflect.New(v.Type()).Elem()

	for i := range v.Len() {
		clonedElem := cc.cloneValue(v.Index(i))
		if clonedElem.IsValid() {
			newArray.Index(i).Set(clonedElem)
		}
	}

	return newArray
}

// cloneInterface creates a deep copy of an interface value.
func (cc *cloneContext) cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	clonedValue := cc.cloneValue(v.Elem())
	if clonedValue.IsValid() {
		newInterface := reflect.New(v.Type()).Elem()
		newInterface.Set(clonedValue)
		return newInterface
	}

	return v
}

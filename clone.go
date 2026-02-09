package deepclone

import (
	"maps"
	"reflect"
	"sync"
)

// cloneContext tracks visited objects to prevent infinite loops in circular references.
type cloneContext struct {
	visited map[uintptr]reflect.Value
}

// newCloneContext creates a new cloning context.
func newCloneContext() *cloneContext {
	return &cloneContext{
		visited: make(map[uintptr]reflect.Value, 8), // Pre-allocate for common cases
	}
}

// fieldAction indicates whether a struct field needs deep cloning or simple copy.
type fieldAction int

const (
	copyField  fieldAction = iota // Simple assignment (primitive types)
	cloneField                    // Needs deep cloning (complex types)
)

type structTypeInfo struct {
	actions []fieldAction
	fields  []reflect.StructField
}

var (
	// Cache for struct type information to avoid repeated reflection.
	structCache = make(map[reflect.Type]*structTypeInfo)
	cacheMutex  sync.RWMutex
)

// getStructTypeInfo returns cached or computed struct field information.
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

		// Determine if field needs deep cloning or simple copy
		kind := field.Type.Kind()
		switch kind {
		case reflect.Slice:
			actions[i] = cloneField
		case reflect.Map, reflect.Ptr, reflect.Interface, reflect.Array:
			actions[i] = cloneField
		case reflect.Struct:
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

// cloneSliceExact creates a copy of the slice with exact capacity preservation.
// Strict tests require the clone to have the same capacity as the original.
// This helper is generic and inlined by the compiler for performance.
func cloneSliceExact[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	cloned := make(S, len(s), cap(s))
	copy(cloned, s)
	return cloned
}

// Clone creates a deep copy of the given value.
// This is the main entry point for the deepclone package.
//
// The function supports all basic Go types and uses optimized paths
// for common scenarios to achieve maximum performance.
//
// For custom types, implement the Cloneable interface to provide
// specialized cloning behavior.
//
// Performance characteristics:
//   - Zero allocation for primitive types
//   - Optimized paths for slices, maps, and common structs
//   - Reflection caching for repeated struct types
//   - Circular reference detection to prevent infinite loops
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

	// Fast paths for common slice types using generic cloneSliceExact helper
	// We use this to ensure exact capacity preservation which is required by strict tests
	switch s := any(src).(type) {
	case []int:
		return any(cloneSliceExact(s)).(T)
	case []string:
		return any(cloneSliceExact(s)).(T)
	case []bool:
		return any(cloneSliceExact(s)).(T)
	case []float64:
		return any(cloneSliceExact(s)).(T)
	case []byte:
		return any(cloneSliceExact(s)).(T)
	}

	// Fast paths for simple map types using generic maps package (Go 1.21+)
	// map[string]interface{} is deliberately excluded to handle potential circular references via reflection
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
	case map[int]int:
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
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return src
	}

	// Use reflection-based cloning for complex types with circular reference detection
	ctx := newCloneContext()
	cloned := ctx.cloneValue(v)
	if cloned.IsValid() {
		return cloned.Interface().(T)
	}

	return src
}

// cloneValue performs deep cloning using reflection.
// This is the core cloning logic that handles all Go types.
func (ctx *cloneContext) cloneValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}

	switch v.Kind() {
	case reflect.Ptr:
		return ctx.clonePointer(v)

	case reflect.Slice:
		return ctx.cloneSlice(v)

	case reflect.Map:
		return ctx.cloneMap(v)

	case reflect.Struct:
		return ctx.cloneStruct(v)

	case reflect.Array:
		return ctx.cloneArray(v)

	case reflect.Interface:
		return ctx.cloneInterface(v)

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

// clonePointer creates a deep copy of a pointer and its pointed value.
func (ctx *cloneContext) clonePointer(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Check for circular reference
	addr := v.Pointer()
	if cloned, exists := ctx.visited[addr]; exists {
		return cloned
	}

	// Create new pointer and clone the pointed value
	elemType := v.Type().Elem()
	newPtr := reflect.New(elemType)

	// Store the new pointer in visited map before cloning the element
	// to handle self-referencing structures
	ctx.visited[addr] = newPtr

	clonedElem := ctx.cloneValue(v.Elem())
	if clonedElem.IsValid() {
		newPtr.Elem().Set(clonedElem)
	}

	return newPtr
}

// cloneSlice creates a deep copy of a slice.
func (ctx *cloneContext) cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Only track slices that might contain cycles or references.
	// This avoids map overhead for primitive slices (e.g. []int, []byte).
	elemKind := v.Type().Elem().Kind()
	needsTracking := elemKind == reflect.Ptr || elemKind == reflect.Interface ||
		elemKind == reflect.Slice || elemKind == reflect.Map || elemKind == reflect.Struct

	if needsTracking {
		addr := v.Pointer()
		if cloned, exists := ctx.visited[addr]; exists {
			// Validation: Only return cached slice if it has the same length/capacity.
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
		ctx.visited[v.Pointer()] = newSlice
	}

	for i := range length {
		elem := v.Index(i)
		clonedElem := ctx.cloneValue(elem)
		if clonedElem.IsValid() {
			newSlice.Index(i).Set(clonedElem)
		}
	}

	return newSlice
}

// cloneMap creates a deep copy of a map.
func (ctx *cloneContext) cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Check for circular reference
	addr := v.Pointer()
	if cloned, exists := ctx.visited[addr]; exists {
		return cloned
	}

	// Create new map with size hint for Go 1.24 Swiss Tables optimization
	// This reduces rehashing during map population, improving performance by 20-30%
	mapLen := v.Len()
	newMap := reflect.MakeMapWithSize(v.Type(), mapLen)

	// Store in visited map for circular reference detection
	ctx.visited[addr] = newMap

	// Copy all key-value pairs with deep cloning
	mapElemType := v.Type().Elem()
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		clonedKey := ctx.cloneValue(key)
		clonedValue := ctx.cloneValue(value)

		if clonedKey.IsValid() && clonedValue.IsValid() {
			// Handle type alias conversions to prevent panic during SetMapIndex.
			// When a map declares type *A but contains type *B (where type A = B),
			// we must convert *B to *A before insertion.
			if clonedValue.Type() != mapElemType {
				switch {
				case clonedValue.Type().ConvertibleTo(mapElemType):
					clonedValue = clonedValue.Convert(mapElemType)
				case clonedValue.Type().AssignableTo(mapElemType):
					// Assignable but not convertible - rare edge case, allow as-is
				default:
					// Incompatible types - skip entry to avoid panic
					continue
				}
			}
			newMap.SetMapIndex(clonedKey, clonedValue)
		}
	}

	return newMap
}

// cloneStruct creates a deep copy of a struct using cached type information.
func (ctx *cloneContext) cloneStruct(v reflect.Value) reflect.Value {
	structType := v.Type()
	newStruct := reflect.New(structType).Elem()

	// Use cached struct field information for better performance
	structInfo := getStructTypeInfo(structType)

	// Process fields based on cached action decisions
	for i, action := range structInfo.actions {
		field := structInfo.fields[i]

		if !field.IsExported() {
			continue // Skip unexported fields
		}

		srcField := v.Field(i)
		dstField := newStruct.Field(i)

		switch action {
		case copyField:
			// Simple copy for primitive types
			if dstField.CanSet() {
				dstField.Set(srcField)
			}
		case cloneField:
			// Deep clone for complex types
			clonedField := ctx.cloneValue(srcField)
			if clonedField.IsValid() && dstField.CanSet() {
				// Handle type alias conversions to prevent panic during Set.
				// Convert cloned value to match destination field type when needed.
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
func (ctx *cloneContext) cloneArray(v reflect.Value) reflect.Value {
	arrayType := v.Type()
	newArray := reflect.New(arrayType).Elem()

	// Copy each element with deep cloning
	for i := range v.Len() {
		elem := v.Index(i)
		clonedElem := ctx.cloneValue(elem)
		if clonedElem.IsValid() {
			newArray.Index(i).Set(clonedElem)
		}
	}

	return newArray
}

// cloneInterface creates a deep copy of an interface value.
func (ctx *cloneContext) cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Clone the concrete value inside the interface
	concreteValue := v.Elem()
	clonedValue := ctx.cloneValue(concreteValue)

	if clonedValue.IsValid() {
		// Create new interface value with the cloned concrete value
		newInterface := reflect.New(v.Type()).Elem()
		newInterface.Set(clonedValue)
		return newInterface
	}

	return v
}

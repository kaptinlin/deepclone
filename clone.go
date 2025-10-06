package deepclone

import (
	"reflect"
	"sync"
)

// cloneContext tracks visited objects to prevent infinite loops in circular references
type cloneContext struct {
	visited map[uintptr]reflect.Value
}

// newCloneContext creates a new cloning context
func newCloneContext() *cloneContext {
	return &cloneContext{
		visited: make(map[uintptr]reflect.Value, 8), // Pre-allocate for common cases
	}
}

// fieldTypeCache caches field action decisions for struct types
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
	// Cache for struct type information to avoid repeated reflection
	structCache = make(map[reflect.Type]*structTypeInfo)
	cacheMutex  sync.RWMutex
)

// getStructTypeInfo returns cached or computed struct field information
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

	for i := 0; i < numFields; i++ {
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
			// All slices need deep cloning (both primitive and complex element types)
			actions[i] = cloneField
		case reflect.Map, reflect.Ptr, reflect.Interface, reflect.Array:
			actions[i] = cloneField
		case reflect.Struct:
			actions[i] = cloneField
		case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Chan, reflect.Func, reflect.String, reflect.UnsafePointer:
			// Primitive types: bool, int*, uint*, float*, complex*, string, chan, func, etc.
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

	// Fast paths for common slice types - avoid reflection for better performance
	switch s := any(src).(type) {
	case []int:
		if s == nil {
			return src
		}
		cloned := make([]int, len(s))
		copy(cloned, s)
		return any(cloned).(T)
	case []string:
		if s == nil {
			return src
		}
		cloned := make([]string, len(s))
		copy(cloned, s)
		return any(cloned).(T)
	case []bool:
		if s == nil {
			return src
		}
		cloned := make([]bool, len(s))
		copy(cloned, s)
		return any(cloned).(T)
	case []float64:
		if s == nil {
			return src
		}
		cloned := make([]float64, len(s))
		copy(cloned, s)
		return any(cloned).(T)
	case []byte:
		if s == nil {
			return src
		}
		cloned := make([]byte, len(s))
		copy(cloned, s)
		return any(cloned).(T)
	}

	// Fast paths for common map types - avoid reflection for better performance
	switch m := any(src).(type) {
	case map[string]int:
		if m == nil {
			return src
		}
		cloned := make(map[string]int, len(m))
		for k, v := range m {
			cloned[k] = v
		}
		return any(cloned).(T)
	case map[string]string:
		if m == nil {
			return src
		}
		cloned := make(map[string]string, len(m))
		for k, v := range m {
			cloned[k] = v
		}
		return any(cloned).(T)
	case map[int]int:
		if m == nil {
			return src
		}
		cloned := make(map[int]int, len(m))
		for k, v := range m {
			cloned[k] = v
		}
		return any(cloned).(T)
	case map[string]interface{}:
		if m == nil {
			return src
		}
		cloned := make(map[string]interface{}, len(m))
		for k, v := range m {
			cloned[k] = Clone(v) // Recursively clone values
		}
		return any(cloned).(T)
	}

	// Fast paths for small structs - avoid reflection overhead for common patterns
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src
	}

	// Check if type implements Cloneable interface FIRST
	// This must come before any fast paths to respect custom cloning behavior
	//
	// The Cloneable interface allows types to implement custom deep cloning logic.
	// When a type implements Cloneable, its Clone() method takes full responsibility
	// for creating a deep copy of the object, including all nested data structures.
	//
	// This provides complete control over the cloning process for complex types
	// that may need special handling, optimization, or custom semantics.
	if cloneable, ok := any(src).(Cloneable); ok {
		if result, ok := cloneable.Clone().(T); ok {
			return result
		}
	}

	// Note: Small struct fast path was removed as it added overhead
	// The cached struct type info path below is more efficient

	// Fast path for nil values without additional reflection
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
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		// Primitive types - return as-is (no allocation needed)
		return v

	case reflect.String:
		// Strings are immutable in Go, so we can return the original
		return v

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
		// Channels cannot be meaningfully cloned, return nil channel of same type
		return reflect.Zero(v.Type())

	case reflect.Func:
		// Functions cannot be cloned, return the original
		return v

	case reflect.Invalid, reflect.UnsafePointer:
		// For invalid types and unsafe pointers, return the original value
		return v

	default:
		// For any other types not explicitly handled, return the original value
		return v
	}
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

	// Check for circular reference (slices can be circular through pointers)
	if v.Len() > 0 && v.Index(0).Kind() == reflect.Ptr {
		addr := v.Pointer()
		if cloned, exists := ctx.visited[addr]; exists {
			return cloned
		}
	}

	length := v.Len()
	capacity := v.Cap()

	// Create new slice with same length and capacity
	newSlice := reflect.MakeSlice(v.Type(), length, capacity)

	// Store in visited map for circular reference detection
	if v.Len() > 0 && v.Index(0).Kind() == reflect.Ptr {
		ctx.visited[v.Pointer()] = newSlice
	}

	// Copy elements with deep cloning
	for i := 0; i < length; i++ {
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

	// Create new map of same type
	newMap := reflect.MakeMap(v.Type())

	// Store in visited map for circular reference detection
	ctx.visited[addr] = newMap

	// Copy all key-value pairs with deep cloning
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		clonedKey := ctx.cloneValue(key)
		clonedValue := ctx.cloneValue(value)

		if clonedKey.IsValid() && clonedValue.IsValid() {
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
	for i := 0; i < v.Len(); i++ {
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

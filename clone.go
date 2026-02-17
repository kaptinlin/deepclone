package deepclone

import (
	"maps"
	"reflect"
	"sync"
)

type fieldAction int

const (
	copyField  fieldAction = iota
	cloneField
)

var (
	// structCache caches struct type information keyed by reflect.Type.
	// Bounded by the number of distinct struct types in the program.
	// Use ResetCache to reclaim memory if needed (e.g., in tests).
	structCache = make(map[reflect.Type]*structTypeInfo)
	cacheMutex  sync.RWMutex
)

// cloneContext tracks visited objects to detect circular references.
type cloneContext struct {
	visited map[uintptr]reflect.Value
}

func newCloneContext() *cloneContext {
	return &cloneContext{
		visited: make(map[uintptr]reflect.Value, 8),
	}
}

type structTypeInfo struct {
	actions []fieldAction
	fields  []reflect.StructField
}

// structInfo returns cached struct field information for the given type,
// computing and caching it on first access.
func structInfo(t reflect.Type) *structTypeInfo {
	cacheMutex.RLock()
	if info, exists := structCache[t]; exists {
		cacheMutex.RUnlock()
		return info
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

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
			actions[i] = copyField
			continue
		}

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

// CacheStats returns the number of struct types currently cached and
// the total number of cached fields across all types.
// CacheStats is safe for concurrent use.
func CacheStats() (entries, fields int) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entries = len(structCache)
	for _, info := range structCache {
		fields += len(info.fields)
	}
	return
}

// ResetCache clears the struct type cache. Subsequent Clone operations
// will re-populate the cache on demand.
// ResetCache is safe for concurrent use.
func ResetCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	clear(structCache)
}

// cloneSliceExact creates a copy of the slice with exact capacity preservation.
func cloneSliceExact[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	cloned := make(S, len(s), cap(s))
	copy(cloned, s)
	return cloned
}

// cloneMapExact creates a shallow copy of the map, preserving nil.
func cloneMapExact[M ~map[K]V, K comparable, V any](m M) M {
	if m == nil {
		return nil
	}
	return maps.Clone(m)
}

// Clone creates a deep copy of the given value, preserving the complete
// object graph including circular references.
//
// Clone uses a hierarchical optimization strategy for maximum performance:
//   - Primitive types (int, string, bool, etc.) return as-is with zero allocation
//   - Common slice types ([]int, []string, []byte, etc.) use optimized generic copy
//   - Common map types (map[string]string, etc.) use maps.Clone
//   - Types implementing Cloneable delegate to their Clone method
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
	// Fast path: primitive types are value types, return as-is.
	switch any(src).(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128,
		string:
		return src
	}

	// Fast path: common slice types using generic copy with exact capacity.
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

	// Fast path: simple map types using maps.Clone.
	// map[string]any is excluded to handle potential circular references via reflection.
	switch m := any(src).(type) {
	case map[string]int:
		return any(cloneMapExact(m)).(T)
	case map[string]string:
		return any(cloneMapExact(m)).(T)
	case map[string]float64:
		return any(cloneMapExact(m)).(T)
	case map[string]bool:
		return any(cloneMapExact(m)).(T)
	case map[int]int:
		return any(cloneMapExact(m)).(T)
	case map[int]string:
		return any(cloneMapExact(m)).(T)
	case map[int]bool:
		return any(cloneMapExact(m)).(T)
	}

	// Reflection-based path for complex types.
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src
	}

	if cloneable, ok := any(src).(Cloneable); ok {
		if result, ok := cloneable.Clone().(T); ok {
			return result
		}
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return src
	}

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
		return reflect.Zero(v.Type())
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Func, reflect.String, reflect.UnsafePointer:
		return v
	}

	return v
}

func (cc *cloneContext) clonePointer(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	addr := v.Pointer()
	if cloned, exists := cc.visited[addr]; exists {
		return cloned
	}

	ptr := reflect.New(v.Type().Elem())

	// Register before recursing to handle self-referencing structures.
	cc.visited[addr] = ptr

	elem := cc.cloneValue(v.Elem())
	if elem.IsValid() {
		ptr.Elem().Set(elem)
	}

	return ptr
}

func (cc *cloneContext) cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Only track slices whose elements can contain cycles.
	elemKind := v.Type().Elem().Kind()
	needsTracking := elemKind == reflect.Pointer || elemKind == reflect.Interface ||
		elemKind == reflect.Slice || elemKind == reflect.Map || elemKind == reflect.Struct

	if needsTracking {
		addr := v.Pointer()
		if cloned, exists := cc.visited[addr]; exists {
			if cloned.Len() == v.Len() && cloned.Cap() == v.Cap() {
				return cloned
			}
		}
	}

	slice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())

	if needsTracking {
		cc.visited[v.Pointer()] = slice
	}

	for i := range v.Len() {
		elem := cc.cloneValue(v.Index(i))
		if elem.IsValid() {
			slice.Index(i).Set(elem)
		}
	}

	return slice
}

func (cc *cloneContext) cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	addr := v.Pointer()
	if cloned, exists := cc.visited[addr]; exists {
		return cloned
	}

	m := reflect.MakeMapWithSize(v.Type(), v.Len())
	cc.visited[addr] = m

	elemType := v.Type().Elem()
	for _, key := range v.MapKeys() {
		k := cc.cloneValue(key)
		val := cc.cloneValue(v.MapIndex(key))

		if !k.IsValid() || !val.IsValid() {
			continue
		}

		// Handle type alias conversions to prevent panic during SetMapIndex.
		if val.Type() != elemType {
			if val.Type().ConvertibleTo(elemType) {
				val = val.Convert(elemType)
			} else if !val.Type().AssignableTo(elemType) {
				continue
			}
		}
		m.SetMapIndex(k, val)
	}

	return m
}

func (cc *cloneContext) cloneStruct(v reflect.Value) reflect.Value {
	s := reflect.New(v.Type()).Elem()
	info := structInfo(v.Type())

	for i, action := range info.actions {
		if !info.fields[i].IsExported() {
			continue
		}

		src := v.Field(i)
		dst := s.Field(i)

		switch action {
		case copyField:
			if dst.CanSet() {
				dst.Set(src)
			}
		case cloneField:
			field := cc.cloneValue(src)
			if field.IsValid() && dst.CanSet() {
				if field.Type() != dst.Type() && field.Type().ConvertibleTo(dst.Type()) {
					field = field.Convert(dst.Type())
				}
				dst.Set(field)
			}
		}
	}

	return s
}

func (cc *cloneContext) cloneArray(v reflect.Value) reflect.Value {
	arr := reflect.New(v.Type()).Elem()

	for i := range v.Len() {
		elem := cc.cloneValue(v.Index(i))
		if elem.IsValid() {
			arr.Index(i).Set(elem)
		}
	}

	return arr
}

func (cc *cloneContext) cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	val := cc.cloneValue(v.Elem())
	if val.IsValid() {
		iface := reflect.New(v.Type()).Elem()
		iface.Set(val)
		return iface
	}

	return v
}

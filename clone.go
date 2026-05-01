package deepclone

import (
	"maps"
	"reflect"
	"sync"
)

type fieldAction int

const (
	copyField fieldAction = iota
	cloneField
)

var (
	// structCache is bounded by the number of distinct struct types seen.
	structCache = make(map[reflect.Type]*structTypeInfo)
	cacheMutex  sync.RWMutex
)

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
}

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

	for i := range numFields {
		field := t.Field(i)

		if field.IsExported() && shouldCloneKind(field.Type.Kind()) {
			actions[i] = cloneField
		}
	}

	info := &structTypeInfo{actions: actions}
	structCache[t] = info
	return info
}

// CacheStats returns the number of cached struct types and fields.
func CacheStats() (entries, fields int) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entries = len(structCache)
	for _, info := range structCache {
		fields += len(info.actions)
	}
	return
}

// ResetCache clears the cached struct metadata.
func ResetCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	clear(structCache)
}

func cloneSliceExact[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	cloned := make(S, len(s), cap(s))
	copy(cloned, s)
	return cloned
}

func shouldCloneKind(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Map || kind == reflect.Pointer ||
		kind == reflect.Interface || kind == reflect.Array || kind == reflect.Struct
}

func sliceCanContainCycles(kind reflect.Kind) bool {
	return kind == reflect.Pointer || kind == reflect.Interface || kind == reflect.Slice ||
		kind == reflect.Map || kind == reflect.Struct
}

func assignableClone(value reflect.Value, target reflect.Type) (reflect.Value, bool) {
	if value.Type() == target || value.Type().AssignableTo(target) {
		return value, true
	}
	if value.Type().ConvertibleTo(target) {
		return value.Convert(target), true
	}
	return reflect.Value{}, false
}

// Clone returns a deep copy of src.
//
// Clone preserves circular references when it uses reflection. Types that
// implement Cloneable control their own cloning behavior.
func Clone[T any](src T) T {
	switch any(src).(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128,
		string:
		return src
	}

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

	// map[string]any is excluded so reflection can preserve circular references.
	switch m := any(src).(type) {
	case map[string]int:
		return any(maps.Clone(m)).(T)
	case map[string]string:
		return any(maps.Clone(m)).(T)
	case map[string]float64:
		return any(maps.Clone(m)).(T)
	case map[string]bool:
		return any(maps.Clone(m)).(T)
	case map[int]int:
		return any(maps.Clone(m)).(T)
	case map[int]string:
		return any(maps.Clone(m)).(T)
	case map[int]bool:
		return any(maps.Clone(m)).(T)
	}

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

	ctx := newCloneContext()
	cloned := ctx.cloneValue(v)
	if cloned.IsValid() {
		return cloned.Interface().(T)
	}

	return src
}

func (c *cloneContext) cloneValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}

	switch v.Kind() {
	case reflect.Pointer:
		return c.clonePointer(v)
	case reflect.Slice:
		return c.cloneSlice(v)
	case reflect.Map:
		return c.cloneMap(v)
	case reflect.Struct:
		return c.cloneStruct(v)
	case reflect.Array:
		return c.cloneArray(v)
	case reflect.Interface:
		return c.cloneInterface(v)
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

func (c *cloneContext) clonePointer(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	addr := v.Pointer()
	if cloned, exists := c.visited[addr]; exists {
		return cloned
	}

	ptr := reflect.New(v.Type().Elem())

	// Register before recursing to handle self-referencing structures.
	c.visited[addr] = ptr

	elem := c.cloneValue(v.Elem())
	if elem.IsValid() {
		ptr.Elem().Set(elem)
	}

	return ptr
}

func (c *cloneContext) cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	// Only track slices whose elements can contain cycles.
	needsTracking := sliceCanContainCycles(v.Type().Elem().Kind())

	if needsTracking {
		addr := v.Pointer()
		if cloned, exists := c.visited[addr]; exists {
			if cloned.Len() == v.Len() && cloned.Cap() == v.Cap() {
				return cloned
			}
		}
	}

	slice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())

	if needsTracking {
		c.visited[v.Pointer()] = slice
	}

	for i := range v.Len() {
		elem := c.cloneValue(v.Index(i))
		if elem.IsValid() {
			slice.Index(i).Set(elem)
		}
	}

	return slice
}

func (c *cloneContext) cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	addr := v.Pointer()
	if cloned, exists := c.visited[addr]; exists {
		return cloned
	}

	m := reflect.MakeMapWithSize(v.Type(), v.Len())
	c.visited[addr] = m

	elemType := v.Type().Elem()
	iter := v.MapRange()
	for iter.Next() {
		k := c.cloneValue(iter.Key())
		val := c.cloneValue(iter.Value())

		if !k.IsValid() || !val.IsValid() {
			continue
		}

		val, ok := assignableClone(val, elemType)
		if !ok {
			continue
		}
		m.SetMapIndex(k, val)
	}

	return m
}

func (c *cloneContext) cloneStruct(v reflect.Value) reflect.Value {
	s := reflect.New(v.Type()).Elem()
	info := structInfo(v.Type())

	for i, action := range info.actions {
		src := v.Field(i)
		dst := s.Field(i)

		switch action {
		case copyField:
			if dst.CanSet() {
				dst.Set(src)
			}
		case cloneField:
			clonedField := c.cloneValue(src)
			if clonedField.IsValid() && dst.CanSet() {
				if clonedField, ok := assignableClone(clonedField, dst.Type()); ok {
					dst.Set(clonedField)
				}
			}
		}
	}

	return s
}

func (c *cloneContext) cloneArray(v reflect.Value) reflect.Value {
	arr := reflect.New(v.Type()).Elem()

	for i := range v.Len() {
		elem := c.cloneValue(v.Index(i))
		if elem.IsValid() {
			arr.Index(i).Set(elem)
		}
	}

	return arr
}

func (c *cloneContext) cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	clonedElem := c.cloneValue(v.Elem())
	if clonedElem.IsValid() {
		iface := reflect.New(v.Type()).Elem()
		iface.Set(clonedElem)
		return iface
	}

	return v
}

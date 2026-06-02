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
	structCache   = make(map[reflect.Type]*structTypeInfo)
	cacheMutex    sync.RWMutex
	cloneableType = reflect.TypeFor[Cloneable]()
)

type visitKind uint8

const (
	visitPointer visitKind = iota
	visitSlice
	visitMap
)

type visitKey struct {
	kind visitKind
	addr uintptr
	typ  reflect.Type
}

type cloneContext struct {
	visited map[visitKey]reflect.Value
}

func newCloneContext() *cloneContext {
	return &cloneContext{
		visited: make(map[visitKey]reflect.Value, 8),
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

	actions := make([]fieldAction, t.NumField())

	for field := range t.Fields() {
		if field.IsExported() && shouldCloneType(field.Type) {
			actions[field.Index[0]] = cloneField
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
		kind == reflect.Interface || kind == reflect.Array || kind == reflect.Struct ||
		kind == reflect.Chan
}

func shouldCloneType(t reflect.Type) bool {
	return shouldCloneKind(t.Kind()) || t.Implements(cloneableType)
}

func sliceCanContainCycles(kind reflect.Kind) bool {
	return kind == reflect.Pointer || kind == reflect.Interface || kind == reflect.Slice ||
		kind == reflect.Map || kind == reflect.Struct || kind == reflect.Array
}

func assignableClone(value reflect.Value, target reflect.Type) (reflect.Value, bool) {
	valueType := value.Type()
	if valueType == target || valueType.AssignableTo(target) {
		return value, true
	}
	if valueType.ConvertibleTo(target) {
		return value.Convert(target), true
	}
	return reflect.Value{}, false
}

func cloneableValue(v reflect.Value) (reflect.Value, bool) {
	if v.Kind() == reflect.Interface || !v.CanInterface() {
		return reflect.Value{}, false
	}

	cloneable, ok := v.Interface().(Cloneable)
	if !ok {
		return reflect.Value{}, false
	}

	result := reflect.ValueOf(cloneable.Clone())
	if !result.IsValid() {
		return reflect.Value{}, false
	}
	if cloned, ok := assignableClone(result, v.Type()); ok {
		return cloned, true
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
	case []byte:
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

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return src
	}

	if cloneable, ok := any(src).(Cloneable); ok {
		result := reflect.ValueOf(cloneable.Clone())
		if result.IsValid() {
			if cloned, ok := assignableClone(result, reflect.TypeFor[T]()); ok {
				return cloned.Interface().(T)
			}
		}
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
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return v
	}
	if cloned, ok := cloneableValue(v); ok {
		return cloned
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

	key := visitKey{kind: visitPointer, addr: v.Pointer(), typ: v.Type()}
	if cloned, exists := c.visited[key]; exists {
		return cloned
	}

	clonedPtr := reflect.New(v.Type().Elem())

	// Register before recursing to handle self-referencing structures.
	c.visited[key] = clonedPtr

	elemValue := v.Elem()
	if elemValue.Kind() == reflect.Struct {
		c.cloneStructInto(elemValue, clonedPtr.Elem())
		return clonedPtr
	}

	elem := c.cloneValue(elemValue)
	if elem.IsValid() {
		clonedPtr.Elem().Set(elem)
	}

	return clonedPtr
}

func (c *cloneContext) cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	needsTracking := sliceCanContainCycles(v.Type().Elem().Kind())
	addr := uintptr(0)

	if needsTracking {
		addr = v.Pointer()
		key := visitKey{kind: visitSlice, addr: addr}
		cloned, exists := c.visited[key]
		if exists && cloned.Len() == v.Len() && cloned.Cap() == v.Cap() {
			return cloned
		}
	}

	clonedSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())

	if needsTracking {
		c.visited[visitKey{kind: visitSlice, addr: addr}] = clonedSlice
	}

	for i := range v.Len() {
		elem := c.cloneValue(v.Index(i))
		if elem.IsValid() {
			clonedSlice.Index(i).Set(elem)
		}
	}

	return clonedSlice
}

func (c *cloneContext) cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	key := visitKey{kind: visitMap, addr: v.Pointer()}
	if cloned, exists := c.visited[key]; exists {
		return cloned
	}

	clonedMap := reflect.MakeMapWithSize(v.Type(), v.Len())
	c.visited[key] = clonedMap

	keyType := v.Type().Key()
	elemType := v.Type().Elem()
	iter := v.MapRange()
	for iter.Next() {
		srcKey := iter.Key()
		srcValue := iter.Value()
		pointerKey, keyIsPointer := pointerVisitKey(srcKey)

		var key reflect.Value
		var value reflect.Value
		if keyIsPointer {
			key = c.cloneValue(srcKey)
			value = c.cloneValue(srcValue)
			key = c.refreshPointerKey(pointerKey, key)
		} else {
			value = c.cloneValue(srcValue)
			key = c.cloneValue(srcKey)
		}

		if !key.IsValid() || !value.IsValid() {
			continue
		}

		key, ok := assignableClone(key, keyType)
		if !ok {
			continue
		}

		value, ok = assignableClone(value, elemType)
		if !ok {
			continue
		}
		clonedMap.SetMapIndex(key, value)
	}

	return clonedMap
}

func (c *cloneContext) refreshPointerKey(key visitKey, clonedKey reflect.Value) reflect.Value {
	if current, exists := c.visited[key]; exists {
		return current
	}
	return clonedKey
}

func pointerVisitKey(v reflect.Value) (visitKey, bool) {
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return visitKey{}, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return visitKey{}, false
	}
	return visitKey{kind: visitPointer, addr: v.Pointer(), typ: v.Type()}, true
}

func (c *cloneContext) cloneStruct(v reflect.Value) reflect.Value {
	clonedStruct := reflect.New(v.Type()).Elem()
	c.cloneStructInto(v, clonedStruct)
	return clonedStruct
}

func (c *cloneContext) registerAddress(src, dst reflect.Value) {
	if src.CanAddr() && dst.CanAddr() {
		addr := src.Addr()
		c.visited[visitKey{kind: visitPointer, addr: addr.Pointer(), typ: addr.Type()}] = dst.Addr()
	}
}

func (c *cloneContext) registerArrayElements(v, clonedArray reflect.Value) {
	for i := range v.Len() {
		src := v.Index(i)
		dst := clonedArray.Index(i)
		c.registerAddress(src, dst)

		switch src.Kind() {
		case reflect.Struct:
			if !src.Type().Implements(cloneableType) {
				c.registerStructFields(src, dst)
			}
		case reflect.Array:
			c.registerArrayElements(src, dst)
		default:
		}
	}
}

func (c *cloneContext) registerStructFields(v, clonedStruct reflect.Value) {
	for i := range v.NumField() {
		fieldInfo := v.Type().Field(i)
		if !fieldInfo.IsExported() {
			continue
		}

		src := v.Field(i)
		dst := clonedStruct.Field(i)
		c.registerAddress(src, dst)

		switch src.Kind() {
		case reflect.Struct:
			if !src.Type().Implements(cloneableType) {
				c.registerStructFields(src, dst)
			}
		case reflect.Array:
			c.registerArrayElements(src, dst)
		default:
		}
	}
}

func (c *cloneContext) cloneStructInto(v, clonedStruct reflect.Value) {
	info := structInfo(v.Type())
	c.registerStructFields(v, clonedStruct)

	for i, action := range info.actions {
		src := v.Field(i)
		dst := clonedStruct.Field(i)

		switch action {
		case copyField:
			if dst.CanSet() {
				dst.Set(src)
			}
		case cloneField:
			if src.Kind() == reflect.Struct && dst.CanSet() && !src.Type().Implements(cloneableType) {
				c.cloneStructInto(src, dst)
				continue
			}

			clonedField := c.cloneValue(src)
			if clonedField.IsValid() && dst.CanSet() {
				if clonedField, ok := assignableClone(clonedField, dst.Type()); ok {
					dst.Set(clonedField)
				}
			}
		}
	}
}

func (c *cloneContext) cloneArray(v reflect.Value) reflect.Value {
	clonedArray := reflect.New(v.Type()).Elem()
	c.registerArrayElements(v, clonedArray)

	for i := range v.Len() {
		elem := c.cloneValue(v.Index(i))
		if elem.IsValid() {
			clonedArray.Index(i).Set(elem)
		}
	}

	return clonedArray
}

func (c *cloneContext) cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	elem := v.Elem()
	if (elem.Kind() != reflect.Pointer || !elem.IsNil()) && v.CanInterface() {
		if cloneable, ok := v.Interface().(Cloneable); ok {
			result := reflect.ValueOf(cloneable.Clone())
			if result.IsValid() {
				if cloned, ok := assignableClone(result, v.Type()); ok {
					iface := reflect.New(v.Type()).Elem()
					iface.Set(cloned)
					return iface
				}
			}
		}
	}

	clonedElem := c.cloneValue(elem)
	if !clonedElem.IsValid() {
		return v
	}

	iface := reflect.New(v.Type()).Elem()
	iface.Set(clonedElem)
	return iface
}

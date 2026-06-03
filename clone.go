package deepclone

import (
	"maps"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
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
	errorType   = reflect.TypeFor[error]()
)

var unsupportedTypes = map[reflect.Type]string{
	reflect.TypeFor[os.File]():        "files cannot be cloned",
	reflect.TypeFor[sync.Cond]():      "sync primitives cannot be cloned",
	reflect.TypeFor[sync.Map]():       "sync primitives cannot be cloned",
	reflect.TypeFor[sync.Mutex]():     "sync primitives cannot be cloned",
	reflect.TypeFor[sync.Once]():      "sync primitives cannot be cloned",
	reflect.TypeFor[sync.Pool]():      "sync primitives cannot be cloned",
	reflect.TypeFor[sync.RWMutex]():   "sync primitives cannot be cloned",
	reflect.TypeFor[sync.WaitGroup](): "sync primitives cannot be cloned",
	reflect.TypeFor[atomic.Bool]():    "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Int32]():   "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Int64]():   "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Uint32]():  "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Uint64]():  "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Uintptr](): "atomic state cannot be cloned",
	reflect.TypeFor[atomic.Value]():   "atomic state cannot be cloned",
}

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
	fields []structFieldInfo
}

type structFieldInfo struct {
	index    int
	name     string
	exported bool
	action   fieldAction
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

	fields := make([]structFieldInfo, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)
		info := structFieldInfo{
			index:    i,
			name:     field.Name,
			exported: field.IsExported(),
			action:   copyField,
		}
		if info.exported && shouldCloneType(field.Type) {
			info.action = cloneField
		}
		fields[i] = info
	}

	info := &structTypeInfo{fields: fields}
	structCache[t] = info
	return info
}

func cacheStats() (entries, fields int) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entries = len(structCache)
	for _, info := range structCache {
		fields += len(info.fields)
	}
	return
}

func resetCache() {
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
		kind == reflect.Chan || kind == reflect.Func || kind == reflect.UnsafePointer
}

func shouldCloneType(t reflect.Type) bool {
	return shouldCloneKind(t.Kind()) || hasCustomCloneType(t)
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

func unsupportedValue(v reflect.Value, path string) error {
	if isNil(v) {
		return nil
	}
	if reason, ok := unsupportedTypeReason(v.Type()); ok {
		return unsupportedError(path, v.Type(), reason)
	}

	switch v.Kind() {
	case reflect.Chan:
		return unsupportedError(path, v.Type(), "channels cannot be cloned")
	case reflect.Func:
		return unsupportedError(path, v.Type(), "functions cannot be cloned")
	case reflect.UnsafePointer:
		return unsupportedError(path, v.Type(), "unsafe pointers cannot be cloned")
	default:
		return nil
	}
}

func unsupportedTypeReason(t reflect.Type) (string, bool) {
	if reason, ok := unsupportedTypes[t]; ok {
		return reason, true
	}
	if t.Kind() == reflect.Pointer {
		return unsupportedTypeReason(t.Elem())
	}
	return "", false
}

func unsupportedUnexportedField(v reflect.Value, path string) error {
	if err := unsupportedValue(v, path); err != nil {
		return err
	}
	if isReferenceLike(v.Kind()) && !isNil(v) {
		return unsupportedError(path, v.Type(), "unexported reference-like fields cannot be cloned")
	}
	return nil
}

func isReferenceLike(kind reflect.Kind) bool {
	return kind == reflect.Pointer || kind == reflect.Slice || kind == reflect.Map ||
		kind == reflect.Interface
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer,
		reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}

func hasCustomCloneType(t reflect.Type) bool {
	_, ok := customCloneMethod(t, t)
	return ok
}

func customCloneMethod(t, target reflect.Type) (reflect.Method, bool) {
	method, ok := t.MethodByName("Clone")
	if !ok {
		return reflect.Method{}, false
	}

	methodType := method.Type
	if methodType.NumIn() != 1 || methodType.NumOut() != 2 {
		return reflect.Method{}, false
	}
	if methodType.Out(1) != errorType {
		return reflect.Method{}, false
	}
	output := methodType.Out(0)
	if output == target || output.AssignableTo(target) || output.ConvertibleTo(target) {
		return method, true
	}
	return reflect.Method{}, false
}

func customCloneValue(v reflect.Value, path string) (reflect.Value, bool, error) {
	if v.Kind() == reflect.Interface || !v.CanInterface() {
		return reflect.Value{}, false, nil
	}
	if _, ok := customCloneMethod(v.Type(), v.Type()); !ok {
		return reflect.Value{}, false, nil
	}

	results := v.MethodByName("Clone").Call(nil)
	if !results[1].IsNil() {
		return reflect.Value{}, true, results[1].Interface().(error)
	}

	cloned, ok := assignableClone(results[0], v.Type())
	if !ok {
		return reflect.Value{}, true, unsupportedError(path, v.Type(), "Cloner returned an incompatible type")
	}
	return cloned, true, nil
}

// Clone returns a deep copy of src.
//
// Clone preserves circular references when it uses reflection. Types that
// implement Cloner[T] control their own cloning behavior.
func Clone[T any](src T) (T, error) {
	switch any(src).(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128,
		string:
		return src, nil
	}

	switch s := any(src).(type) {
	case []int:
		return any(cloneSliceExact(s)).(T), nil
	case []int8:
		return any(cloneSliceExact(s)).(T), nil
	case []int16:
		return any(cloneSliceExact(s)).(T), nil
	case []int32:
		return any(cloneSliceExact(s)).(T), nil
	case []int64:
		return any(cloneSliceExact(s)).(T), nil
	case []uint:
		return any(cloneSliceExact(s)).(T), nil
	case []byte:
		return any(cloneSliceExact(s)).(T), nil
	case []uint16:
		return any(cloneSliceExact(s)).(T), nil
	case []uint32:
		return any(cloneSliceExact(s)).(T), nil
	case []uint64:
		return any(cloneSliceExact(s)).(T), nil
	case []float32:
		return any(cloneSliceExact(s)).(T), nil
	case []float64:
		return any(cloneSliceExact(s)).(T), nil
	case []string:
		return any(cloneSliceExact(s)).(T), nil
	case []bool:
		return any(cloneSliceExact(s)).(T), nil
	}

	// map[string]any is excluded so reflection can preserve circular references.
	switch m := any(src).(type) {
	case map[string]int:
		return any(maps.Clone(m)).(T), nil
	case map[string]string:
		return any(maps.Clone(m)).(T), nil
	case map[string]float64:
		return any(maps.Clone(m)).(T), nil
	case map[string]bool:
		return any(maps.Clone(m)).(T), nil
	case map[int]int:
		return any(maps.Clone(m)).(T), nil
	case map[int]string:
		return any(maps.Clone(m)).(T), nil
	case map[int]bool:
		return any(maps.Clone(m)).(T), nil
	}

	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return src, nil
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return src, nil
	}

	if cloner, ok := any(src).(Cloner[T]); ok {
		return cloner.Clone()
	}

	ctx := newCloneContext()
	cloned, err := ctx.cloneValue(v, "$")
	if err != nil {
		var zero T
		return zero, err
	}
	if cloned.IsValid() {
		return cloned.Interface().(T), nil
	}

	return src, nil
}

// MustClone returns a deep copy of src or panics if src cannot be cloned.
func MustClone[T any](src T) T {
	cloned, err := Clone(src)
	if err != nil {
		panic(err)
	}
	return cloned
}

func (c *cloneContext) cloneValue(v reflect.Value, path string) (reflect.Value, error) {
	if !v.IsValid() {
		return reflect.Value{}, nil
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return v, nil
	}
	if cloned, ok, err := customCloneValue(v, path); ok || err != nil {
		return cloned, err
	}
	if err := unsupportedValue(v, path); err != nil {
		return reflect.Value{}, err
	}

	switch v.Kind() {
	case reflect.Pointer:
		return c.clonePointer(v, path)
	case reflect.Slice:
		return c.cloneSlice(v, path)
	case reflect.Map:
		return c.cloneMap(v, path)
	case reflect.Struct:
		return c.cloneStruct(v, path)
	case reflect.Array:
		return c.cloneArray(v, path)
	case reflect.Interface:
		return c.cloneInterface(v, path)
	case reflect.Chan:
		return v, nil
	case reflect.Func:
		return v, nil
	case reflect.UnsafePointer:
		return v, nil
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return v, nil
	}

	return v, nil
}

func (c *cloneContext) clonePointer(v reflect.Value, path string) (reflect.Value, error) {
	if v.IsNil() {
		return v, nil
	}

	key := visitKey{kind: visitPointer, addr: v.Pointer(), typ: v.Type()}
	if cloned, exists := c.visited[key]; exists {
		return cloned, nil
	}

	clonedPtr := reflect.New(v.Type().Elem())

	// Register before recursing to handle self-referencing structures.
	c.visited[key] = clonedPtr

	elemValue := v.Elem()
	if elemValue.Kind() == reflect.Struct {
		clonedPtr.Elem().Set(elemValue)
		if err := c.cloneStructInto(elemValue, clonedPtr.Elem(), path); err != nil {
			return reflect.Value{}, err
		}
		return clonedPtr, nil
	}

	elem, err := c.cloneValue(elemValue, path)
	if err != nil {
		return reflect.Value{}, err
	}
	if elem.IsValid() {
		clonedPtr.Elem().Set(elem)
	}

	return clonedPtr, nil
}

func (c *cloneContext) cloneSlice(v reflect.Value, path string) (reflect.Value, error) {
	if v.IsNil() {
		return v, nil
	}

	needsTracking := sliceCanContainCycles(v.Type().Elem().Kind())
	addr := uintptr(0)

	if needsTracking {
		addr = v.Pointer()
		key := visitKey{kind: visitSlice, addr: addr, typ: v.Type()}
		cloned, exists := c.visited[key]
		if exists && cloned.Len() == v.Len() && cloned.Cap() == v.Cap() {
			return cloned, nil
		}
	}

	clonedSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())

	if needsTracking {
		c.visited[visitKey{kind: visitSlice, addr: addr, typ: v.Type()}] = clonedSlice
	}

	for i := range v.Len() {
		elem, err := c.cloneValue(v.Index(i), indexPath(path, i))
		if err != nil {
			return reflect.Value{}, err
		}
		if elem.IsValid() {
			clonedSlice.Index(i).Set(elem)
		}
	}

	return clonedSlice, nil
}

func (c *cloneContext) cloneMap(v reflect.Value, path string) (reflect.Value, error) {
	if v.IsNil() {
		return v, nil
	}

	key := visitKey{kind: visitMap, addr: v.Pointer(), typ: v.Type()}
	if cloned, exists := c.visited[key]; exists {
		return cloned, nil
	}

	clonedMap := reflect.MakeMapWithSize(v.Type(), v.Len())
	c.visited[key] = clonedMap

	keyType := v.Type().Key()
	elemType := v.Type().Elem()
	iter := v.MapRange()
	for iter.Next() {
		srcKey := iter.Key()
		srcValue := iter.Value()

		value, err := c.cloneValue(srcValue, mapValuePath(path, srcKey))
		if err != nil {
			return reflect.Value{}, err
		}
		key, err := c.cloneValue(srcKey, mapKeyPath(path, srcKey))
		if err != nil {
			return reflect.Value{}, err
		}

		if !key.IsValid() || !value.IsValid() {
			return reflect.Value{}, unsupportedError(path, v.Type(), "map key or value cloned to an invalid value")
		}

		key, ok := assignableClone(key, keyType)
		if !ok {
			return reflect.Value{}, unsupportedError(mapKeyPath(path, srcKey), srcKey.Type(), "cloned map key is not assignable to the map key type")
		}

		value, ok = assignableClone(value, elemType)
		if !ok {
			return reflect.Value{}, unsupportedError(mapValuePath(path, srcKey), srcValue.Type(), "cloned map value is not assignable to the map value type")
		}
		clonedMap.SetMapIndex(key, value)
	}

	return clonedMap, nil
}

func (c *cloneContext) cloneStruct(v reflect.Value, path string) (reflect.Value, error) {
	clonedStruct := reflect.New(v.Type()).Elem()
	clonedStruct.Set(v)
	if err := c.cloneStructInto(v, clonedStruct, path); err != nil {
		return reflect.Value{}, err
	}
	return clonedStruct, nil
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
			if !hasCustomCloneType(src.Type()) {
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
			if !hasCustomCloneType(src.Type()) {
				c.registerStructFields(src, dst)
			}
		case reflect.Array:
			c.registerArrayElements(src, dst)
		default:
		}
	}
}

func (c *cloneContext) cloneStructInto(v, clonedStruct reflect.Value, path string) error {
	info := structInfo(v.Type())
	c.registerStructFields(v, clonedStruct)

	for _, field := range info.fields {
		src := v.Field(field.index)
		dst := clonedStruct.Field(field.index)
		fieldNamePath := fieldPath(path, field.name)
		if field.exported {
			if err := unsupportedValue(src, fieldNamePath); err != nil {
				return err
			}
		} else {
			if err := unsupportedUnexportedField(src, fieldNamePath); err != nil {
				return err
			}
		}

		switch field.action {
		case copyField:
			if dst.CanSet() {
				dst.Set(src)
			}
		case cloneField:
			if src.Kind() == reflect.Struct && dst.CanSet() && !hasCustomCloneType(src.Type()) {
				if err := c.cloneStructInto(src, dst, fieldNamePath); err != nil {
					return err
				}
				continue
			}
			if src.Kind() == reflect.Array && dst.CanSet() {
				if err := c.cloneArrayInto(src, dst, fieldNamePath); err != nil {
					return err
				}
				continue
			}

			clonedField, err := c.cloneValue(src, fieldNamePath)
			if err != nil {
				return err
			}
			if clonedField.IsValid() && dst.CanSet() {
				if clonedField, ok := assignableClone(clonedField, dst.Type()); ok {
					dst.Set(clonedField)
				} else {
					return unsupportedError(fieldNamePath, src.Type(), "cloned field is not assignable to the field type")
				}
			}
		}
	}
	return nil
}

func (c *cloneContext) cloneArray(v reflect.Value, path string) (reflect.Value, error) {
	clonedArray := reflect.New(v.Type()).Elem()
	c.registerArrayElements(v, clonedArray)

	if err := c.cloneArrayInto(v, clonedArray, path); err != nil {
		return reflect.Value{}, err
	}
	return clonedArray, nil
}

func (c *cloneContext) cloneArrayInto(v, clonedArray reflect.Value, path string) error {
	for i := range v.Len() {
		elem, err := c.cloneValue(v.Index(i), indexPath(path, i))
		if err != nil {
			return err
		}
		if elem.IsValid() {
			clonedArray.Index(i).Set(elem)
		}
	}
	return nil
}

func (c *cloneContext) cloneInterface(v reflect.Value, path string) (reflect.Value, error) {
	if v.IsNil() {
		return v, nil
	}

	clonedElem, err := c.cloneValue(v.Elem(), path)
	if err != nil {
		return reflect.Value{}, err
	}
	if !clonedElem.IsValid() {
		return v, nil
	}

	iface := reflect.New(v.Type()).Elem()
	iface.Set(clonedElem)
	return iface, nil
}

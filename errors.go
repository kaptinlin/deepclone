package deepclone

import (
	"fmt"
	"reflect"
	"strconv"
)

// UnsupportedError reports a value that cannot be honestly deep-cloned.
type UnsupportedError struct {
	Path   string
	Type   reflect.Type
	Reason string
}

func (e *UnsupportedError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("deepclone: unsupported value at %s (%s): %s", e.Path, e.Type, e.Reason)
}

func unsupportedError(path string, typ reflect.Type, reason string) error {
	if path == "" {
		path = "$"
	}
	return &UnsupportedError{
		Path:   path,
		Type:   typ,
		Reason: reason,
	}
}

func fieldPath(path, name string) string {
	if path == "" {
		path = "$"
	}
	if name == "" {
		return path + ".?"
	}
	return path + "." + name
}

func indexPath(path string, index int) string {
	if path == "" {
		path = "$"
	}
	return path + "[" + strconv.Itoa(index) + "]"
}

func mapKeyPath(path string, key reflect.Value) string {
	if path == "" {
		path = "$"
	}
	return path + "[" + mapKeyLabel(key) + "]"
}

func mapValuePath(path string, key reflect.Value) string {
	return mapKeyPath(path, key)
}

func mapKeyLabel(key reflect.Value) string {
	for key.IsValid() && key.Kind() == reflect.Interface {
		if key.IsNil() {
			return "nil"
		}
		key = key.Elem()
	}

	if !key.IsValid() {
		return "key"
	}

	switch key.Kind() {
	case reflect.String:
		return strconv.Quote(key.String())
	case reflect.Bool:
		return strconv.FormatBool(key.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(key.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(key.Uint(), 10)
	default:
		return "key"
	}
}

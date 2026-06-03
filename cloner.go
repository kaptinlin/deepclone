package deepclone

// Cloner lets a type define its own deep-cloning behavior.
//
// Clone must return a copy that can be used independently of the original.
// Circular reference detection does not apply inside custom Clone methods.
type Cloner[T any] interface {
	// Clone returns a copy that can be used independently of the original.
	Clone() (T, error)
}

package deepclone

// Cloneable lets a type define its own deep-cloning behavior.
//
// Clone must return a copy that can be used independently of the original.
// Circular reference detection does not apply inside custom Clone methods.
type Cloneable interface {
	// Clone returns a copy that can be used independently of the original.
	Clone() any
}

// Package deepclone copies Go values whose state can be represented as
// memory-owned data.
//
// Clone returns a deep copy or an error when a value cannot be honestly cloned.
// MustClone is the convenience form for values that are known to be supported.
//
// Reflection cloning preserves supported object graphs, including circular
// references. Nil pointers, slices, maps, interfaces, channels, functions, and
// unsafe pointers keep their nil meaning. Non-nil channels, functions, and
// unsafe pointers are rejected because they represent runtime identity or
// execution capability rather than ordinary memory-owned data.
//
// The package does not use unsafe to read or write unexported fields. Reflection
// cloning preserves value-like unexported fields by shallow-copying the struct
// first, but rejects unexported reference-like state that it cannot safely
// deep-clone. Types with private invariants or resource ownership should
// implement Cloner[T] and define their own behavior.
package deepclone

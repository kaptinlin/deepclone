// Package deepclone provides deep cloning with fast paths for common value,
// slice, and map types and reflection-based cloning for the rest.
//
// Clone preserves object graphs, including circular references, when it falls
// back to reflection. Types that implement Cloneable provide their own cloning
// behavior and are responsible for handling cycles inside that method.
package deepclone

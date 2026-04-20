// Package main demonstrates custom cloning behavior with the deepclone library.
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

// Counter shows custom Cloneable behavior.
type Counter struct {
	// Value is incremented by Clone.
	Value int
	// Name gets a suffix in Clone.
	Name string
}

// Clone returns a modified copy to show custom clone hooks.
func (c Counter) Clone() any {
	return Counter{
		Value: c.Value + 1,
		Name:  c.Name + "_copy",
	}
}

// Person uses the default deep-cloning path.
type Person struct {
	// Name is copied by the default clone path.
	Name string
	// Age is copied by the default clone path.
	Age int
}

func main() {
	fmt.Println("=== Custom Cloneable Interface Example ===")

	fmt.Println("1. Custom Counter cloning:")
	original := Counter{Value: 10, Name: "main"}
	cloned := deepclone.Clone(original)

	fmt.Printf("Original: Value=%d, Name=%s\n", original.Value, original.Name)
	fmt.Printf("Cloned:   Value=%d, Name=%s\n\n", cloned.Value, cloned.Name)

	fmt.Println("2. Default deep cloning:")
	person := Person{Name: "Alice", Age: 30}
	clonedPerson := deepclone.Clone(person)

	person.Name = "Alice Modified"
	fmt.Printf("Original: Name=%s, Age=%d\n", person.Name, person.Age)
	fmt.Printf("Cloned:   Name=%s, Age=%d\n", clonedPerson.Name, clonedPerson.Age)
}

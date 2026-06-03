// Package main demonstrates custom cloning behavior with the deepclone library.
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

type counter struct {
	Value int
	Name  string
}

func (c counter) Clone() (counter, error) {
	return counter{
		Value: c.Value + 1,
		Name:  c.Name + "_copy",
	}, nil
}

type person struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("=== Custom Cloner Interface Example ===")

	fmt.Println("1. Custom Counter cloning:")
	original := counter{Value: 10, Name: "main"}
	cloned := deepclone.MustClone(original)

	fmt.Printf("Original: Value=%d, Name=%s\n", original.Value, original.Name)
	fmt.Printf("Cloned:   Value=%d, Name=%s\n\n", cloned.Value, cloned.Name)

	fmt.Println("2. Default deep cloning:")
	person := person{Name: "Alice", Age: 30}
	clonedPerson := deepclone.MustClone(person)

	person.Name = "Alice Modified"
	fmt.Printf("Original: Name=%s, Age=%d\n", person.Name, person.Age)
	fmt.Printf("Cloned:   Name=%s, Age=%d\n", clonedPerson.Name, clonedPerson.Age)
}

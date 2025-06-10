package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

type Counter struct {
	Value int
	Name  string
}

func (c Counter) Clone() any {
	return Counter{
		Value: c.Value + 1,
		Name:  c.Name + "_copy",
	}
}

type Person struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("=== Custom Cloneable Interface Example ===")

	// Custom cloning behavior
	fmt.Println("1. Custom Counter cloning:")
	original := Counter{Value: 10, Name: "main"}
	cloned := deepclone.Clone(original)

	fmt.Printf("Original: Value=%d, Name=%s\n", original.Value, original.Name)
	fmt.Printf("Cloned:   Value=%d, Name=%s\n\n", cloned.Value, cloned.Name)

	// Default deep cloning
	fmt.Println("2. Default deep cloning:")
	person := Person{Name: "Alice", Age: 30}
	clonedPerson := deepclone.Clone(person)

	person.Name = "Alice Modified"
	fmt.Printf("Original: Name=%s, Age=%d\n", person.Name, person.Age)
	fmt.Printf("Cloned:   Name=%s, Age=%d\n", clonedPerson.Name, clonedPerson.Age)
}

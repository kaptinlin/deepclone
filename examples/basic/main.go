// Package main demonstrates basic usage of the deepclone library.
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

// User is used in the struct-cloning example.
type User struct {
	// Name is the user's display name.
	Name string
	// Age is the user's age.
	Age int
	// Friends lists related users.
	Friends []string
	// Metadata stores additional example attributes.
	Metadata map[string]any
}

// CustomData shows how Cloneable can customize cloning.
type CustomData struct {
	// Value is cloned without modification.
	Value string
	// Count is incremented by Clone.
	Count int
}

// Clone increments Count to show custom clone behavior.
func (c CustomData) Clone() any {
	return CustomData{
		Value: c.Value,
		Count: c.Count + 1, // Increment count on clone
	}
}

func main() {
	fmt.Println("=== DeepClone Basic Examples ===")

	fmt.Println("1. Basic types:")
	original := 42
	cloned := deepclone.Clone(original)
	fmt.Printf("Original: %d, Cloned: %d\n\n", original, cloned)

	fmt.Println("2. Slices:")
	originalSlice := []int{1, 2, 3, 4, 5}
	clonedSlice := deepclone.Clone(originalSlice)

	originalSlice[0] = 999
	fmt.Printf("Original: %v\n", originalSlice)
	fmt.Printf("Cloned:   %v\n\n", clonedSlice)

	fmt.Println("3. Maps:")
	originalMap := map[string]int{
		"apple":  1,
		"banana": 2,
		"orange": 3,
	}
	clonedMap := deepclone.Clone(originalMap)

	originalMap["grape"] = 4
	fmt.Printf("Original: %v\n", originalMap)
	fmt.Printf("Cloned:   %v\n\n", clonedMap)

	fmt.Println("4. Pointers:")
	value := 100
	originalPtr := &value
	clonedPtr := deepclone.Clone(originalPtr)

	*originalPtr = 200
	fmt.Printf("Original: %d (addr: %p)\n", *originalPtr, originalPtr)
	fmt.Printf("Cloned:   %d (addr: %p)\n\n", *clonedPtr, clonedPtr)

	fmt.Println("5. Structs:")
	originalUser := User{
		Name:    "John Doe",
		Age:     30,
		Friends: []string{"Alice", "Bob"},
		Metadata: map[string]any{
			"score":  85.5,
			"active": true,
		},
	}
	clonedUser := deepclone.Clone(originalUser)

	originalUser.Name = "Jane Doe"
	originalUser.Friends[0] = "Charlie"
	originalUser.Metadata["score"] = 95.0

	fmt.Printf("Original: %+v\n", originalUser)
	fmt.Printf("Cloned:   %+v\n\n", clonedUser)

	fmt.Println("6. Custom Cloneable interface:")
	originalCustom := CustomData{Value: "test", Count: 1}
	clonedCustom := deepclone.Clone(originalCustom)

	fmt.Printf("Original: %+v\n", originalCustom)
	fmt.Printf("Cloned:   %+v\n", clonedCustom)
}

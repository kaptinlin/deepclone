// Package main demonstrates basic usage of the deepclone library.
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

type user struct {
	Name     string
	Age      int
	Friends  []string
	Metadata map[string]any
}

type customData struct {
	Value string
	Count int
}

func (c customData) Clone() any {
	return customData{
		Value: c.Value,
		Count: c.Count + 1,
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
	originalUser := user{
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
	originalCustom := customData{Value: "test", Count: 1}
	clonedCustom := deepclone.Clone(originalCustom)

	fmt.Printf("Original: %+v\n", originalCustom)
	fmt.Printf("Cloned:   %+v\n", clonedCustom)
}

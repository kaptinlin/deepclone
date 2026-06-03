package deepclone_test

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

func ExampleClone() {
	original := map[string][]int{
		"scores": {90, 85, 77},
	}
	cloned, err := deepclone.Clone(original)
	if err != nil {
		panic(err)
	}

	// Modify the clone — original is unaffected
	cloned["scores"][0] = 100

	fmt.Println("original:", original["scores"])
	fmt.Println("cloned:  ", cloned["scores"])
	// Output:
	// original: [90 85 77]
	// cloned:   [100 85 77]
}

func ExampleClone_struct() {
	type Address struct {
		City  string
		State string
	}
	type Person struct {
		Name    string
		Age     int
		Address *Address
	}

	original := Person{
		Name: "Alice",
		Age:  30,
		Address: &Address{
			City:  "Portland",
			State: "OR",
		},
	}
	cloned := deepclone.MustClone(original)

	// Modify the clone's nested pointer — original is unaffected
	cloned.Address.City = "Seattle"
	cloned.Address.State = "WA"

	fmt.Println("original:", original.Address.City, original.Address.State)
	fmt.Println("cloned:  ", cloned.Address.City, cloned.Address.State)
	// Output:
	// original: Portland OR
	// cloned:   Seattle WA
}

func ExampleClone_slice() {
	original := []string{"a", "b", "c"}
	cloned := deepclone.MustClone(original)

	cloned[0] = "z"

	fmt.Println("original:", original)
	fmt.Println("cloned:  ", cloned)
	// Output:
	// original: [a b c]
	// cloned:   [z b c]
}

func ExampleClone_nil() {
	var original []int
	cloned := deepclone.MustClone(original)

	fmt.Println("nil preserved:", cloned == nil)
	// Output:
	// nil preserved: true
}

// Document is a type that implements the Cloner interface
// to provide custom deep cloning behavior.
type Document struct {
	Title string
	Tags  []string
}

func (d Document) Clone() (Document, error) {
	tags, err := deepclone.Clone(d.Tags)
	if err != nil {
		return Document{}, err
	}
	return Document{
		Title: d.Title,
		Tags:  tags,
	}, nil
}

func ExampleCloner() {
	original := Document{
		Title: "Guide",
		Tags:  []string{"go", "clone"},
	}
	cloned := deepclone.MustClone(original)

	cloned.Tags[0] = "rust"

	fmt.Println("original:", original.Tags)
	fmt.Println("cloned:  ", cloned.Tags)
	// Output:
	// original: [go clone]
	// cloned:   [rust clone]
}

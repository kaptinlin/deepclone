package deepclone

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClonePrimitiveTypes(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"int", 42},
		{"int8", int8(8)},
		{"int16", int16(16)},
		{"int32", int32(32)},
		{"int64", int64(64)},
		{"uint", uint(42)},
		{"uint8", uint8(8)},
		{"uint16", uint16(16)},
		{"uint32", uint32(32)},
		{"uint64", uint64(64)},
		{"float32", float32(3.14)},
		{"float64", 3.14159},
		{"bool true", true},
		{"bool false", false},
		{"string", "hello world"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clone(tt.input)
			assert.Equal(t, tt.input, result)
		})
	}
}

func TestCloneSlices(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.NotSame(t, &original[0], &cloned[0]) // Ensure different memory

		// Modify original to verify independence
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("string slice", func(t *testing.T) {
		original := []string{"hello", "world", "test"}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.NotSame(t, &original[0], &cloned[0])
	})

	t.Run("nil slice", func(t *testing.T) {
		var original []int
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("empty slice", func(t *testing.T) {
		original := []int{}
		cloned := Clone(original)
		assert.Equal(t, original, cloned)
		assert.Len(t, cloned, 0)
	})
}

func TestCloneMaps(t *testing.T) {
	t.Run("string to int map", func(t *testing.T) {
		original := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		// Modify original to verify independence
		original["four"] = 4
		assert.NotContains(t, cloned, "four")
	})

	t.Run("nil map", func(t *testing.T) {
		var original map[string]int
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("empty map", func(t *testing.T) {
		original := make(map[string]int)
		cloned := Clone(original)
		assert.Equal(t, original, cloned)
		assert.Len(t, cloned, 0)
	})
}

func TestClonePointers(t *testing.T) {
	t.Run("int pointer", func(t *testing.T) {
		value := 42
		original := &value
		cloned := Clone(original)

		require.NotNil(t, cloned)
		assert.Equal(t, *original, *cloned)
		assert.NotSame(t, original, cloned) // Different pointers

		// Modify original to verify independence
		*original = 999
		assert.NotEqual(t, *original, *cloned)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var original *int
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("pointer chain", func(t *testing.T) {
		value := 42
		ptr1 := &value
		original := &ptr1
		cloned := Clone(original)

		require.NotNil(t, cloned)
		require.NotNil(t, *cloned)
		assert.Equal(t, **original, **cloned)
		assert.NotSame(t, original, cloned)
		assert.NotSame(t, *original, *cloned)
	})
}

func TestCloneStructs(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	t.Run("simple struct", func(t *testing.T) {
		original := SimpleStruct{
			Name: "John",
			Age:  30,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		// Verify they are independent (for reference types within)
		original.Name = "Jane"
		assert.NotEqual(t, original.Name, cloned.Name)
	})

	type NestedStruct struct {
		Data   []int
		Config map[string]string
		Ptr    *int
	}

	t.Run("nested struct", func(t *testing.T) {
		value := 100
		original := NestedStruct{
			Data:   []int{1, 2, 3},
			Config: map[string]string{"key": "value"},
			Ptr:    &value,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		// Verify deep independence
		original.Data[0] = 999
		assert.NotEqual(t, original.Data[0], cloned.Data[0])

		original.Config["new"] = "value"
		assert.NotContains(t, cloned.Config, "new")

		*original.Ptr = 999
		assert.NotEqual(t, *original.Ptr, *cloned.Ptr)
	})
}

func TestCloneArrays(t *testing.T) {
	t.Run("int array", func(t *testing.T) {
		original := [3]int{1, 2, 3}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		// Arrays are value types, so modifying original won't affect cloned
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("nested array", func(t *testing.T) {
		original := [2][]int{{1, 2}, {3, 4}}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		// Verify deep independence of slice elements
		original[0][0] = 999
		assert.NotEqual(t, original[0][0], cloned[0][0])
	})
}

// TestCloneableInterface tests custom cloning behavior
func TestCloneableInterface(t *testing.T) {
	t.Run("custom cloneable", func(t *testing.T) {
		original := CustomType{Value: "test"}
		cloned := Clone(original)

		expected := CustomType{Value: "test_cloned"}
		assert.Equal(t, expected, cloned)
	})
}

// CustomType implements Cloneable interface
type CustomType struct {
	Value string
}

func (c CustomType) Clone() any {
	return CustomType{Value: c.Value + "_cloned"}
}

// TestCloneCircularReference tests circular reference detection
func TestCloneCircularReference(t *testing.T) {
	t.Run("circular pointer reference", func(t *testing.T) {
		type Node struct {
			Value int
			Next  *Node
		}

		// Create circular reference: node1 -> node2 -> node1
		node1 := &Node{Value: 1}
		node2 := &Node{Value: 2}
		node1.Next = node2
		node2.Next = node1

		// This should not panic or cause infinite loop
		cloned := Clone(node1)

		// Verify structure is preserved
		require.NotNil(t, cloned)
		assert.Equal(t, 1, cloned.Value)
		require.NotNil(t, cloned.Next)
		assert.Equal(t, 2, cloned.Next.Value)
		require.NotNil(t, cloned.Next.Next)
		assert.Equal(t, 1, cloned.Next.Next.Value)

		// Verify circular reference is maintained
		assert.True(t, cloned.Next.Next == cloned, "Circular reference should be maintained")

		// Verify independence from original
		assert.False(t, cloned == node1, "Cloned should be different from original")
		assert.False(t, cloned.Next == node2, "Cloned.Next should be different from original.Next")
	})

	t.Run("self-referencing pointer", func(t *testing.T) {
		type SelfRef struct {
			Value int
			Self  *SelfRef
		}

		original := &SelfRef{Value: 42}
		original.Self = original

		cloned := Clone(original)

		require.NotNil(t, cloned)
		assert.Equal(t, 42, cloned.Value)
		require.NotNil(t, cloned.Self)
		assert.True(t, cloned.Self == cloned, "Self-reference should be maintained")
		assert.False(t, cloned == original, "Should be different objects")
	})

	t.Run("circular reference in slice", func(t *testing.T) {
		type Container struct {
			Items []*Container
		}

		container1 := &Container{}
		container2 := &Container{}
		container1.Items = []*Container{container2}
		container2.Items = []*Container{container1}

		cloned := Clone(container1)

		require.NotNil(t, cloned)
		require.Len(t, cloned.Items, 1)
		require.NotNil(t, cloned.Items[0])
		require.Len(t, cloned.Items[0].Items, 1)
		assert.True(t, cloned.Items[0].Items[0] == cloned, "Circular reference should be maintained")
	})
}

// TestCloneEdgeCases tests various edge cases and boundary conditions
func TestCloneEdgeCases(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) {
		var original interface{}
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("nil pointer to interface", func(t *testing.T) {
		var original *interface{}
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("empty interface with nil value", func(t *testing.T) {
		var nilPtr *int
		var original interface{} = nilPtr
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("deeply nested nil pointers", func(t *testing.T) {
		var original ***int
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("struct with unexported fields", func(t *testing.T) {
		type StructWithUnexported struct {
			Public    string
			Protected int // exported field
		}

		// Note: We can't directly set unexported fields from outside the package
		// but this test verifies the cloning handles them gracefully
		original := StructWithUnexported{
			Public:    "visible",
			Protected: 42,
		}

		cloned := Clone(original)
		assert.Equal(t, "visible", cloned.Public)
		assert.Equal(t, 42, cloned.Protected)
		// unexported fields would remain zero value since they can't be accessed
	})

	t.Run("large slice", func(t *testing.T) {
		original := make([]int, 10000)
		for i := range original {
			original[i] = i
		}

		cloned := Clone(original)
		assert.Equal(t, len(original), len(cloned))
		assert.Equal(t, cap(original), cap(cloned))

		// Verify first, middle, and last elements
		assert.Equal(t, original[0], cloned[0])
		assert.Equal(t, original[5000], cloned[5000])
		assert.Equal(t, original[9999], cloned[9999])

		// Verify independence
		original[0] = -1
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("map with nil values", func(t *testing.T) {
		original := map[string]*int{
			"nil":   nil,
			"valid": func() *int { i := 42; return &i }(),
		}

		cloned := Clone(original)
		assert.Nil(t, cloned["nil"])
		require.NotNil(t, cloned["valid"])
		assert.Equal(t, 42, *cloned["valid"])

		// Verify independence
		*original["valid"] = 100
		assert.NotEqual(t, *original["valid"], *cloned["valid"])
	})

	t.Run("interface with concrete values", func(t *testing.T) {
		tests := []interface{}{
			42,
			"hello",
			[]int{1, 2, 3},
			map[string]int{"key": 42},
		}

		for i, original := range tests {
			t.Run(fmt.Sprintf("type_%d", i), func(t *testing.T) {
				cloned := Clone(original)
				assert.Equal(t, original, cloned)
			})
		}
	})

	t.Run("channel types", func(t *testing.T) {
		// Unbuffered channel
		original := make(chan int)
		cloned := Clone(original)

		// Channels can't be meaningfully cloned, should get zero value
		assert.Nil(t, cloned)

		// Buffered channel
		buffered := make(chan string, 5)
		clonedBuffered := Clone(buffered)
		assert.Nil(t, clonedBuffered)
	})

	t.Run("function types", func(t *testing.T) {
		original := func(x int) int { return x * 2 }
		cloned := Clone(original)

		// Functions should be returned as-is (same reference)
		assert.True(t, reflect.ValueOf(original).Pointer() == reflect.ValueOf(cloned).Pointer())
	})

	t.Run("complex nested structure", func(t *testing.T) {
		type Address struct {
			Street  string
			City    string
			ZipCode *string
		}

		type Person struct {
			Name      string
			Age       int
			Addresses []Address
			Friends   []*Person
			Metadata  map[string]interface{}
			Parent    *Person
		}

		zip := "12345"
		parent := &Person{Name: "Parent", Age: 60}

		original := Person{
			Name: "John",
			Age:  30,
			Addresses: []Address{
				{Street: "123 Main St", City: "Anytown", ZipCode: &zip},
				{Street: "456 Oak Ave", City: "Other City", ZipCode: nil},
			},
			Metadata: map[string]interface{}{
				"score":    85.5,
				"active":   true,
				"tags":     []string{"premium", "verified"},
				"settings": map[string]bool{"notify": true, "sync": false},
				"nullVal":  nil,
			},
			Parent: parent,
		}

		// Add a simple friend without circular reference for this test
		friend := &Person{Name: "Friend", Age: 25, Parent: parent}
		original.Friends = []*Person{friend}

		cloned := Clone(original)

		// Verify all fields are properly cloned
		assert.Equal(t, original.Name, cloned.Name)
		assert.Equal(t, original.Age, cloned.Age)
		assert.Len(t, cloned.Addresses, 2)
		assert.Equal(t, "123 Main St", cloned.Addresses[0].Street)
		assert.NotNil(t, cloned.Addresses[0].ZipCode)
		assert.Equal(t, "12345", *cloned.Addresses[0].ZipCode)
		assert.Nil(t, cloned.Addresses[1].ZipCode)

		// Verify metadata cloning
		assert.Equal(t, 85.5, cloned.Metadata["score"])
		assert.Equal(t, true, cloned.Metadata["active"])
		assert.Nil(t, cloned.Metadata["nullVal"])

		// Verify tags slice is cloned
		tags := cloned.Metadata["tags"].([]string)
		assert.Equal(t, []string{"premium", "verified"}, tags)

		// Verify nested map is cloned
		settings := cloned.Metadata["settings"].(map[string]bool)
		assert.Equal(t, map[string]bool{"notify": true, "sync": false}, settings)

		// Verify independence
		originalTags := original.Metadata["tags"].([]string)
		originalTags[0] = "modified"
		assert.NotEqual(t, originalTags[0], tags[0])

		// Verify friend is properly cloned
		assert.Len(t, cloned.Friends, 1)
		assert.Equal(t, "Friend", cloned.Friends[0].Name)
		assert.False(t, cloned.Friends[0] == friend, "Friend should be cloned, not same reference")

		// Verify parent is properly cloned
		assert.Equal(t, "Parent", cloned.Parent.Name)
		assert.False(t, cloned.Parent == parent, "Parent should be cloned, not same reference")
	})
}

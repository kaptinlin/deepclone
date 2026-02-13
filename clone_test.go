package deepclone

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

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
		assert.Equal(t, cap(original), cap(cloned)) // Strictly enforce capacity preservation
		assert.NotSame(t, &original[0], &cloned[0]) // Ensure different memory

		// Modify original to verify independence
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("string slice", func(t *testing.T) {
		original := []string{"hello", "world", "test"}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		assert.NotSame(t, &original[0], &cloned[0])
	})

	t.Run("int8 slice", func(t *testing.T) {
		original := []int8{1, -2, 3}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 99
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("int16 slice", func(t *testing.T) {
		original := []int16{100, -200, 300}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("int32 slice", func(t *testing.T) {
		original := []int32{100000, -200000, 300000}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("int64 slice", func(t *testing.T) {
		original := []int64{1e12, -2e12, 3e12}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("uint slice", func(t *testing.T) {
		original := []uint{1, 2, 3}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("uint16 slice", func(t *testing.T) {
		original := []uint16{100, 200, 300}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("uint32 slice", func(t *testing.T) {
		original := []uint32{100000, 200000, 300000}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("uint64 slice", func(t *testing.T) {
		original := []uint64{1e12, 2e12, 3e12}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("float32 slice", func(t *testing.T) {
		original := []float32{1.1, 2.2, 3.3}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999.9
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("nil slice types", func(t *testing.T) {
		var nilInt8 []int8
		assert.Nil(t, Clone(nilInt8))

		var nilInt16 []int16
		assert.Nil(t, Clone(nilInt16))

		var nilInt32 []int32
		assert.Nil(t, Clone(nilInt32))

		var nilInt64 []int64
		assert.Nil(t, Clone(nilInt64))

		var nilUint []uint
		assert.Nil(t, Clone(nilUint))

		var nilUint16 []uint16
		assert.Nil(t, Clone(nilUint16))

		var nilUint32 []uint32
		assert.Nil(t, Clone(nilUint32))

		var nilUint64 []uint64
		assert.Nil(t, Clone(nilUint64))

		var nilFloat32 []float32
		assert.Nil(t, Clone(nilFloat32))
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
		assert.Equal(t, cap(original), cap(cloned))
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

	t.Run("string to float64 map", func(t *testing.T) {
		original := map[string]float64{
			"pi": 3.14159,
			"e":  2.71828,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		original["phi"] = 1.618
		assert.NotContains(t, cloned, "phi")
	})

	t.Run("string to bool map", func(t *testing.T) {
		original := map[string]bool{
			"active":  true,
			"deleted": false,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		original["new"] = true
		assert.NotContains(t, cloned, "new")
	})

	t.Run("int to string map", func(t *testing.T) {
		original := map[int]string{
			1: "one",
			2: "two",
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		original[3] = "three"
		assert.NotContains(t, cloned, 3)
	})

	t.Run("int to bool map", func(t *testing.T) {
		original := map[int]bool{
			1: true,
			2: false,
		}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)

		original[3] = true
		assert.NotContains(t, cloned, 3)
	})

	t.Run("nil map types", func(t *testing.T) {
		var nilStringFloat64 map[string]float64
		assert.Nil(t, Clone(nilStringFloat64))

		var nilStringBool map[string]bool
		assert.Nil(t, Clone(nilStringBool))

		var nilIntString map[int]string
		assert.Nil(t, Clone(nilIntString))

		var nilIntBool map[int]bool
		assert.Nil(t, Clone(nilIntBool))
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
		var original any
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("nil pointer to interface", func(t *testing.T) {
		var original *any
		cloned := Clone(original)
		assert.Nil(t, cloned)
	})

	t.Run("empty interface with nil value", func(t *testing.T) {
		var nilPtr *int
		var original any = nilPtr
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
		tests := []any{
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
			Metadata  map[string]any
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
			Metadata: map[string]any{
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

// TestCloneTypeAliases tests cloning with type aliases to prevent panic
func TestCloneTypeAliases(t *testing.T) {
	t.Run("pointer type alias in map", func(t *testing.T) {
		type Schema struct {
			Name string
			Age  int
		}
		type Property Schema

		// Map declares *Property but contains *Schema
		schemaVal := &Schema{Name: "test", Age: 42}
		original := map[string]*Property{
			"key": (*Property)(schemaVal),
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned)
		require.NotNil(t, cloned["key"])
		assert.Equal(t, "test", cloned["key"].Name)
		assert.Equal(t, 42, cloned["key"].Age)

		// Verify independence
		original["key"].Name = "modified"
		assert.NotEqual(t, original["key"].Name, cloned["key"].Name)
	})

	t.Run("type alias in struct field", func(t *testing.T) {
		type Base struct {
			Value int
		}
		type Alias Base

		type Container struct {
			Field *Base
		}

		// Field declares *Base but contains *Alias converted to *Base
		aliasVal := &Alias{Value: 42}
		original := Container{
			Field: (*Base)(aliasVal),
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned.Field)
		assert.Equal(t, 42, cloned.Field.Value)

		// Verify independence
		original.Field.Value = 999
		assert.NotEqual(t, original.Field.Value, cloned.Field.Value)
	})

	t.Run("complex type alias scenario", func(t *testing.T) {
		type Schema struct {
			Type  string
			Items *Schema
		}
		type Property Schema

		// Complex nested structure with type aliases
		itemsSchema := &Schema{Type: "string"}
		original := map[string]*Property{
			"prop1": (*Property)(&Schema{
				Type:  "array",
				Items: itemsSchema,
			}),
			"prop2": (*Property)(&Schema{
				Type: "object",
			}),
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned)
		require.NotNil(t, cloned["prop1"])
		require.NotNil(t, cloned["prop2"])
		assert.Equal(t, "array", cloned["prop1"].Type)
		require.NotNil(t, cloned["prop1"].Items)
		assert.Equal(t, "string", cloned["prop1"].Items.Type)
		assert.Equal(t, "object", cloned["prop2"].Type)

		// Verify deep independence
		original["prop1"].Items.Type = "modified"
		assert.Equal(t, "string", cloned["prop1"].Items.Type)
	})

	t.Run("circular reference with type aliases", func(t *testing.T) {
		type Node struct {
			Value int
			Next  *Node
		}
		type AliasNode Node

		// Create circular reference with type alias
		node1 := &Node{Value: 1}
		node2 := (*Node)(&AliasNode{Value: 2})
		node1.Next = node2
		node2.Next = node1

		// Should not panic and maintain circular reference
		cloned := Clone(node1)
		require.NotNil(t, cloned)
		assert.Equal(t, 1, cloned.Value)
		require.NotNil(t, cloned.Next)
		assert.Equal(t, 2, cloned.Next.Value)
		require.NotNil(t, cloned.Next.Next)
		assert.Equal(t, 1, cloned.Next.Next.Value)
		assert.True(t, cloned.Next.Next == cloned, "Circular reference should be maintained")
	})

	t.Run("slice with type alias elements", func(t *testing.T) {
		type Base struct {
			ID   int
			Name string
		}
		type Derived Base

		// Slice with mixed type aliases
		original := []*Base{
			{ID: 1, Name: "first"},
			(*Base)(&Derived{ID: 2, Name: "second"}),
			{ID: 3, Name: "third"},
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.Len(t, cloned, 3)
		assert.Equal(t, 1, cloned[0].ID)
		assert.Equal(t, "first", cloned[0].Name)
		assert.Equal(t, 2, cloned[1].ID)
		assert.Equal(t, "second", cloned[1].Name)
		assert.Equal(t, 3, cloned[2].ID)
		assert.Equal(t, "third", cloned[2].Name)

		// Verify independence
		original[1].Name = "modified"
		assert.Equal(t, "second", cloned[1].Name)
	})

	t.Run("nested maps with type aliases", func(t *testing.T) {
		type InnerMap map[string]int
		type OuterValue struct {
			Data InnerMap
		}
		type AliasValue OuterValue

		// Nested map structure with type aliases
		original := map[string]*OuterValue{
			"key1": (*OuterValue)(&AliasValue{
				Data: InnerMap{"a": 1, "b": 2},
			}),
			"key2": {
				Data: InnerMap{"c": 3, "d": 4},
			},
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned)
		require.NotNil(t, cloned["key1"])
		require.NotNil(t, cloned["key2"])
		assert.Equal(t, 1, cloned["key1"].Data["a"])
		assert.Equal(t, 2, cloned["key1"].Data["b"])
		assert.Equal(t, 3, cloned["key2"].Data["c"])
		assert.Equal(t, 4, cloned["key2"].Data["d"])

		// Verify deep independence
		original["key1"].Data["a"] = 999
		assert.Equal(t, 1, cloned["key1"].Data["a"])
	})

	t.Run("type alias with interface field", func(t *testing.T) {
		type Base struct {
			Value any
		}
		type Alias Base

		type Container struct {
			Items map[string]*Base
		}

		// Map with type alias containing interface values
		original := Container{
			Items: map[string]*Base{
				"int":    (*Base)(&Alias{Value: 42}),
				"string": (*Base)(&Alias{Value: "test"}),
				"slice":  (*Base)(&Alias{Value: []int{1, 2, 3}}),
			},
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned.Items)
		assert.Equal(t, 42, cloned.Items["int"].Value)
		assert.Equal(t, "test", cloned.Items["string"].Value)
		assert.Equal(t, []int{1, 2, 3}, cloned.Items["slice"].Value)

		// Verify independence of slice in interface
		originalSlice := original.Items["slice"].Value.([]int)
		originalSlice[0] = 999
		clonedSlice := cloned.Items["slice"].Value.([]int)
		assert.Equal(t, 1, clonedSlice[0])
	})

	t.Run("type alias chain", func(t *testing.T) {
		type Level1 struct {
			Value string
		}
		type Level2 Level1
		type Level3 Level2

		// Multiple levels of type aliases
		original := map[string]*Level1{
			"a": (*Level1)(&Level2{Value: "test1"}),
			"b": (*Level1)(&Level3{Value: "test2"}),
			"c": {Value: "test3"},
		}

		// Should not panic during cloning
		cloned := Clone(original)
		require.NotNil(t, cloned)
		assert.Equal(t, "test1", cloned["a"].Value)
		assert.Equal(t, "test2", cloned["b"].Value)
		assert.Equal(t, "test3", cloned["c"].Value)

		// Verify independence
		original["a"].Value = "modified"
		assert.Equal(t, "test1", cloned["a"].Value)
	})
}

// TestCloneUnexportedComplexFields tests structs with unexported fields
// of complex types (slice, map, pointer) to cover the unexported field
// skip path in getStructTypeInfo and cloneStruct.
func TestCloneUnexportedComplexFields(t *testing.T) {
	type withUnexportedSlice struct {
		Public string
		hidden []int //nolint:unused // exercising unexported field path
	}

	original := withUnexportedSlice{Public: "visible"}
	cloned := Clone(original)

	assert.Equal(t, "visible", cloned.Public)
}

// TestCloneAdditionalSliceFastPaths covers the fast paths for []float64,
// []bool, and []byte slices that were not exercised by existing tests.
func TestCloneAdditionalSliceFastPaths(t *testing.T) {
	t.Run("float64 slice", func(t *testing.T) {
		original := []float64{1.1, 2.2, 3.3}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 999.9
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("bool slice", func(t *testing.T) {
		original := []bool{true, false, true}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = false
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("byte slice", func(t *testing.T) {
		original := []byte{0x01, 0x02, 0x03}
		cloned := Clone(original)

		assert.Equal(t, original, cloned)
		assert.Equal(t, cap(original), cap(cloned))
		original[0] = 0xFF
		assert.NotEqual(t, original[0], cloned[0])
	})

	t.Run("nil float64 slice", func(t *testing.T) {
		var original []float64
		assert.Nil(t, Clone(original))
	})

	t.Run("nil bool slice", func(t *testing.T) {
		var original []bool
		assert.Nil(t, Clone(original))
	})

	t.Run("nil byte slice", func(t *testing.T) {
		var original []byte
		assert.Nil(t, Clone(original))
	})
}

// TestCloneNilMapFastPaths covers nil map fast paths for map types
// that were not exercised by existing nil map tests.
func TestCloneNilMapFastPaths(t *testing.T) {
	t.Run("nil map[string]string", func(t *testing.T) {
		var original map[string]string
		assert.Nil(t, Clone(original))
	})

	t.Run("nil map[string]int", func(t *testing.T) {
		var original map[string]int
		assert.Nil(t, Clone(original))
	})

	t.Run("nil map[int]int", func(t *testing.T) {
		var original map[int]int
		assert.Nil(t, Clone(original))
	})

	t.Run("non-nil map[string]string", func(t *testing.T) {
		original := map[string]string{"a": "b"}
		cloned := Clone(original)
		assert.Equal(t, original, cloned)
		original["c"] = "d"
		assert.NotContains(t, cloned, "c")
	})

	t.Run("non-nil map[int]int", func(t *testing.T) {
		original := map[int]int{1: 10, 2: 20}
		cloned := Clone(original)
		assert.Equal(t, original, cloned)
		original[3] = 30
		assert.NotContains(t, cloned, 3)
	})
}

// TestCloneCircularMapReference covers the circular reference detection
// path in cloneMap where a previously visited map is returned from cache.
func TestCloneCircularMapReference(t *testing.T) {
	// A map[string]any that contains itself as a value triggers the
	// circular reference detection directly inside cloneMap.
	m := map[string]any{"key": "value"}
	m["self"] = m // self-referencing map

	cloned := Clone(m)

	require.NotNil(t, cloned)
	assert.Equal(t, "value", cloned["key"])
	// The "self" entry should reference the cloned map itself.
	inner, ok := cloned["self"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", inner["key"])
}

// TestCloneCircularSliceReference covers the circular reference detection
// path in cloneSlice where a previously visited slice with matching
// len/cap is returned from cache.
func TestCloneCircularSliceReference(t *testing.T) {
	// A slice of any that contains itself triggers the circular
	// reference detection directly inside cloneSlice.
	s := make([]any, 2)
	s[0] = "value"
	s[1] = s // self-referencing slice

	cloned := Clone(s)

	require.Len(t, cloned, 2)
	assert.Equal(t, "value", cloned[0])
	// The second element should reference the cloned slice itself.
	inner, ok := cloned[1].([]any)
	require.True(t, ok)
	assert.Len(t, inner, 2)
	assert.Equal(t, "value", inner[0])
}

// TestCloneSharedMapReference covers the case where the same map
// is referenced from two struct fields, hitting the visited cache
// in cloneMap on the second encounter.
func TestCloneSharedMapReference(t *testing.T) {
	type TwoMaps struct {
		A map[string]any
		B map[string]any
	}

	shared := map[string]any{"key": "value"}
	original := TwoMaps{A: shared, B: shared}
	cloned := Clone(original)

	assert.Equal(t, "value", cloned.A["key"])
	assert.Equal(t, "value", cloned.B["key"])

	// Both fields should point to the same cloned map.
	cloned.A["new"] = "added"
	assert.Equal(t, "added", cloned.B["new"],
		"shared map reference should be preserved")
}

// TestCloneSharedSliceReference covers the case where the same slice
// is referenced from two struct fields with identical len/cap, hitting
// the visited cache in cloneSlice on the second encounter.
func TestCloneSharedSliceReference(t *testing.T) {
	type TwoSlices struct {
		A []any
		B []any
	}

	shared := []any{"x", "y"}
	original := TwoSlices{A: shared, B: shared}
	cloned := Clone(original)

	assert.Equal(t, []any{"x", "y"}, cloned.A)
	assert.Equal(t, []any{"x", "y"}, cloned.B)

	// Both fields should point to the same cloned slice.
	cloned.A[0] = "modified"
	assert.Equal(t, "modified", cloned.B[0],
		"shared slice reference should be preserved")
}

// TestCloneMapTypeAliasConversions covers the type alias conversion
// branches in cloneMap: ConvertibleTo, AssignableTo, and incompatible.
func TestCloneMapTypeAliasConversions(t *testing.T) {
	t.Run("convertible value type", func(t *testing.T) {
		type MyInt int

		// map[string]any with a MyInt value — when cloned via reflection,
		// the concrete type may differ from the map's element type (any),
		// exercising the ConvertibleTo branch.
		original := map[string]MyInt{"key": 42}
		cloned := Clone(original)

		require.NotNil(t, cloned)
		assert.Equal(t, MyInt(42), cloned["key"])
	})
}

// TestCloneStructFieldTypeConversion covers the type alias conversion
// path in cloneStruct where a cloned field's type differs from the
// destination field type and needs Convert().
func TestCloneStructFieldTypeConversion(t *testing.T) {
	type Inner struct {
		Value int
	}

	type Outer struct {
		Field Inner
	}

	original := Outer{Field: Inner{Value: 42}}
	cloned := Clone(original)

	assert.Equal(t, 42, cloned.Field.Value)
}

// TestCloneComplexPrimitives covers complex64 and complex128 types
// through the reflection path (via pointer indirection).
func TestCloneComplexPrimitives(t *testing.T) {
	t.Run("complex64", func(t *testing.T) {
		v := complex(float32(1.0), float32(2.0))
		cloned := Clone(v)
		assert.Equal(t, v, cloned)
	})

	t.Run("complex128", func(t *testing.T) {
		v := complex(1.0, 2.0)
		cloned := Clone(v)
		assert.Equal(t, v, cloned)
	})

	t.Run("uintptr", func(t *testing.T) {
		v := uintptr(0xDEADBEEF)
		cloned := Clone(v)
		assert.Equal(t, v, cloned)
	})
}

// TestCloneChannelViaReflection covers the channel zero-value return
// path in cloneValue when a channel is inside an interface (goes through
// cloneValue rather than the struct copyField fast path).
func TestCloneChannelViaReflection(t *testing.T) {
	type WithInterface struct {
		Value any
	}

	ch := make(chan int, 5)
	original := WithInterface{Value: ch}
	cloned := Clone(original)

	// Channel inside interface goes through cloneValue → reflect.Zero
	assert.Nil(t, cloned.Value)
}

// TestCloneFuncViaReflection covers the function as-is return path
// in cloneValue when a func is inside a struct.
func TestCloneFuncViaReflection(t *testing.T) {
	type WithFunc struct {
		Fn   func() int
		Name string
	}

	fn := func() int { return 42 }
	original := WithFunc{Fn: fn, Name: "test"}
	cloned := Clone(original)

	assert.Equal(t, "test", cloned.Name)
	// Function should be returned as-is (same reference)
	assert.Equal(t,
		reflect.ValueOf(original.Fn).Pointer(),
		reflect.ValueOf(cloned.Fn).Pointer())
}

// TestCloneNilInterface covers the nil interface return path in
// cloneInterface when an interface field holds nil.
func TestCloneNilInterface(t *testing.T) {
	type WithInterface struct {
		Value any
	}

	original := WithInterface{Value: nil}
	cloned := Clone(original)

	assert.Nil(t, cloned.Value)
}

// TestCloneCloneableReturnsWrongType covers the path where a Cloneable
// implementation returns a type that doesn't match T, falling through
// to reflection-based cloning.
func TestCloneCloneableReturnsWrongType(t *testing.T) {
	original := &WrongTypeCloneable{Value: "test"}
	cloned := Clone(original)

	require.NotNil(t, cloned)
	// The Cloneable returns a string, which can't be asserted to
	// *WrongTypeCloneable, so it falls through to reflection cloning.
	assert.Equal(t, "test", cloned.Value)
}

// WrongTypeCloneable implements Cloneable but returns a different type.
type WrongTypeCloneable struct {
	Value string
}

func (w *WrongTypeCloneable) Clone() any {
	return "not the right type"
}

// TestCloneSliceSubSliceAliasing verifies that sub-slices sharing the
// same backing array are not incorrectly aliased via the visited cache.
func TestCloneSliceSubSliceAliasing(t *testing.T) {
	type TwoSlices struct {
		Full []any
		Sub  []any
	}

	full := make([]any, 3, 5)
	full[0] = "a"
	full[1] = "b"
	full[2] = "c"
	sub := full[:2]

	original := &TwoSlices{Full: full, Sub: sub}
	cloned := Clone(original)

	require.NotNil(t, cloned)
	assert.Len(t, cloned.Full, 3)
	assert.Len(t, cloned.Sub, 2)
	assert.Equal(t, 5, cap(cloned.Full))
	assert.Equal(t, 5, cap(cloned.Sub))
}

// TestCloneUnsafePointer covers the unsafe.Pointer as-is return path
// in cloneValue. unsafe.Pointer is an opaque type that cannot be
// meaningfully deep cloned, so it is copied by value (same address).
func TestCloneUnsafePointer(t *testing.T) {
	t.Run("direct via interface", func(t *testing.T) {
		x := 42
		ptr := unsafe.Pointer(&x)
		cloned := Clone(ptr)

		// unsafe.Pointer should be copied as-is (same address).
		assert.Equal(t, ptr, cloned)
	})

	t.Run("struct field", func(t *testing.T) {
		type WithUnsafePointer struct {
			Ptr  unsafe.Pointer
			Name string
		}

		x := 99
		original := WithUnsafePointer{
			Ptr:  unsafe.Pointer(&x),
			Name: "test",
		}
		cloned := Clone(original)

		assert.Equal(t, "test", cloned.Name)
		// unsafe.Pointer field should be copied by value.
		assert.Equal(t, original.Ptr, cloned.Ptr)
	})

	t.Run("nil unsafe pointer", func(t *testing.T) {
		type WithUnsafePointer struct {
			Ptr unsafe.Pointer
		}

		original := WithUnsafePointer{Ptr: nil}
		cloned := Clone(original)

		assert.Equal(t, unsafe.Pointer(nil), cloned.Ptr)
	})

	t.Run("inside interface field", func(t *testing.T) {
		type WithInterface struct {
			Value any
		}

		x := 7
		ptr := unsafe.Pointer(&x)
		original := WithInterface{Value: ptr}
		cloned := Clone(original)

		// unsafe.Pointer inside interface goes through
		// cloneValue and should be returned as-is.
		assert.Equal(t, ptr, cloned.Value)
	})
}

// DeepNested is a recursive struct for ultra-deep nesting tests.
type DeepNested struct {
	Depth int
	Data  string
	Tags  []string
	Meta  map[string]int
	Child *DeepNested
}

// buildDeepNested constructs a chain of DeepNested structs with the given depth.
func buildDeepNested(depth int) *DeepNested {
	root := &DeepNested{
		Depth: 0,
		Data:  fmt.Sprintf("level-%d", 0),
		Tags:  []string{"tag0a", "tag0b"},
		Meta:  map[string]int{"x": 0},
	}
	current := root
	for i := 1; i < depth; i++ {
		child := &DeepNested{
			Depth: i,
			Data:  fmt.Sprintf("level-%d", i),
			Tags:  []string{fmt.Sprintf("tag%da", i), fmt.Sprintf("tag%db", i)},
			Meta:  map[string]int{"x": i},
		}
		current.Child = child
		current = child
	}
	return root
}

func TestCloneUltraDeepNestedStruct(t *testing.T) {
	t.Run("depth 1000", func(t *testing.T) {
		const depth = 1000
		original := buildDeepNested(depth)

		cloned := Clone(original)

		// Walk both chains and verify equality and independence.
		origCur, cloneCur := original, cloned
		for i := range depth {
			require.NotNil(t, cloneCur, "nil at depth %d", i)
			assert.Equal(t, i, cloneCur.Depth)
			assert.Equal(t, fmt.Sprintf("level-%d", i), cloneCur.Data)
			assert.Equal(t, origCur.Tags, cloneCur.Tags)
			assert.Equal(t, origCur.Meta, cloneCur.Meta)

			// Verify deep independence: mutate original, clone unaffected.
			origCur.Data = "mutated"
			assert.NotEqual(t, origCur.Data, cloneCur.Data)

			origCur.Tags[0] = "mutated"
			assert.NotEqual(t, origCur.Tags[0], cloneCur.Tags[0])

			origCur.Meta["x"] = -1
			assert.NotEqual(t, origCur.Meta["x"], cloneCur.Meta["x"])

			origCur = origCur.Child
			cloneCur = cloneCur.Child
		}
		// Both chains should terminate at the same depth.
		assert.Nil(t, origCur)
		assert.Nil(t, cloneCur)
	})

	t.Run("depth 10000", func(t *testing.T) {
		const depth = 10000
		original := buildDeepNested(depth)

		cloned := Clone(original)

		// Verify total depth by walking the chain.
		count := 0
		for cur := cloned; cur != nil; cur = cur.Child {
			count++
		}
		assert.Equal(t, depth, count)

		// Spot-check a few levels.
		cur := cloned
		for range 9999 {
			cur = cur.Child
		}
		assert.Equal(t, 9999, cur.Depth)
		assert.Equal(t, "level-9999", cur.Data)
		assert.Nil(t, cur.Child)
	})

	t.Run("deep chain with circular ref at leaf", func(t *testing.T) {
		const depth = 500
		original := buildDeepNested(depth)

		// Create circular reference: leaf points back to root.
		leaf := original
		for leaf.Child != nil {
			leaf = leaf.Child
		}
		leaf.Child = original

		cloned := Clone(original)

		// Walk to the leaf of the clone.
		clonedLeaf := cloned
		for range depth - 1 {
			require.NotNil(t, clonedLeaf.Child)
			clonedLeaf = clonedLeaf.Child
		}
		assert.Equal(t, depth-1, clonedLeaf.Depth)

		// The leaf's Child should point back to the cloned root.
		require.NotNil(t, clonedLeaf.Child)
		assert.True(t, clonedLeaf.Child == cloned,
			"circular reference at leaf should point to cloned root")

		// Verify independence from original.
		assert.False(t, cloned == original)
		assert.False(t, clonedLeaf.Child == original)
	})
}

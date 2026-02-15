package deepclone

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stressNode is a linked-list node for circular reference stress tests.
type stressNode struct {
	ID       int
	Name     string
	Tags     []string
	Meta     map[string]int
	Children []*stressNode
	Next     *stressNode
}

// stressCloneable implements Cloneable for concurrent testing.
type stressCloneable struct {
	Value int
	Data  []byte
}

func (s stressCloneable) Clone() any {
	data := make([]byte, len(s.Data))
	copy(data, s.Data)
	return stressCloneable{Value: s.Value, Data: data}
}

// TestConcurrentCloneStructs stress-tests concurrent cloning of
// structs with slices and maps to verify data independence.
func TestConcurrentCloneStructs(t *testing.T) {
	const goroutines = 100
	const iterations = 200

	type Config struct {
		Host    string
		Port    int
		Tags    []string
		Options map[string]string
	}

	original := Config{
		Host: "localhost",
		Port: 8080,
		Tags: []string{"prod", "us-east", "primary"},
		Options: map[string]string{
			"timeout": "30s",
			"retries": "3",
		},
	}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				cloned := Clone(original)
				assert.Equal(t, original, cloned)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneSlices stress-tests concurrent cloning of
// typed slices through both fast paths and reflection paths.
func TestConcurrentCloneSlices(t *testing.T) {
	const goroutines = 100
	const iterations = 500

	intSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	strSlice := []string{"a", "b", "c", "d", "e"}
	anySlice := []any{1, "two", 3.0, true, nil}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				c := Clone(intSlice)
				assert.Equal(t, intSlice, c)
				assert.Equal(t, cap(intSlice), cap(c))
			}
		})
		wg.Go(func() {
			for range iterations {
				c := Clone(strSlice)
				assert.Equal(t, strSlice, c)
			}
		})
		wg.Go(func() {
			for range iterations {
				c := Clone(anySlice)
				assert.Equal(t, anySlice, c)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneMaps stress-tests concurrent cloning of
// typed maps through both fast paths and reflection paths.
func TestConcurrentCloneMaps(t *testing.T) {
	const goroutines = 100
	const iterations = 500

	strMap := map[string]string{
		"key1": "val1", "key2": "val2", "key3": "val3",
	}
	intMap := map[string]int{
		"a": 1, "b": 2, "c": 3, "d": 4,
	}
	nestedMap := map[string]any{
		"slice": []int{1, 2, 3},
		"map":   map[string]int{"x": 10},
		"str":   "hello",
	}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				c := Clone(strMap)
				assert.Equal(t, strMap, c)
			}
		})
		wg.Go(func() {
			for range iterations {
				c := Clone(intMap)
				assert.Equal(t, intMap, c)
			}
		})
		wg.Go(func() {
			for range iterations {
				c := Clone(nestedMap)
				assert.Equal(t, nestedMap, c)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneCircularRef stress-tests concurrent cloning
// of structures with circular references to verify no infinite
// loops or data races occur.
func TestConcurrentCloneCircularRef(t *testing.T) {
	const goroutines = 100
	const iterations = 100

	a := &stressNode{ID: 1, Name: "a", Tags: []string{"root"}}
	b := &stressNode{ID: 2, Name: "b", Meta: map[string]int{"x": 1}}
	c := &stressNode{ID: 3, Name: "c"}
	a.Next = b
	b.Next = c
	c.Next = a // circular: a -> b -> c -> a
	a.Children = []*stressNode{b, c}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				cloned := Clone(a)
				assert.Equal(t, a.ID, cloned.ID)
				assert.Equal(t, a.Name, cloned.Name)
				assert.Equal(t, a.Tags, cloned.Tags)
				assert.Equal(t, b.ID, cloned.Next.ID)
				assert.Equal(t, c.ID, cloned.Next.Next.ID)
				// Verify circular ref is preserved.
				assert.Same(t, cloned, cloned.Next.Next.Next)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneCloneable stress-tests concurrent cloning of
// types implementing the Cloneable interface.
func TestConcurrentCloneCloneable(t *testing.T) {
	const goroutines = 100
	const iterations = 500

	original := stressCloneable{
		Value: 42,
		Data:  []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				cloned := Clone(original)
				assert.Equal(t, original.Value, cloned.Value)
				assert.Equal(t, original.Data, cloned.Data)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneMixedTypes stress-tests concurrent cloning of
// many different types simultaneously to exercise all code paths
// (fast paths, Cloneable, reflection) under contention.
func TestConcurrentCloneMixedTypes(t *testing.T) {
	const goroutines = 50
	const iterations = 200

	type Nested struct {
		Inner *Nested
		Value int
		Data  []byte
	}

	ptr := 42
	sources := []struct {
		name string
		fn   func()
	}{
		{"int", func() {
			c := Clone(12345)
			assert.Equal(t, 12345, c)
		}},
		{"string", func() {
			c := Clone("concurrent")
			assert.Equal(t, "concurrent", c)
		}},
		{"pointer", func() {
			c := Clone(&ptr)
			assert.Equal(t, ptr, *c)
			assert.NotSame(t, &ptr, c)
		}},
		{"int_slice", func() {
			s := []int{10, 20, 30}
			c := Clone(s)
			assert.Equal(t, s, c)
		}},
		{"string_map", func() {
			m := map[string]string{"k": "v"}
			c := Clone(m)
			assert.Equal(t, m, c)
		}},
		{"nested_struct", func() {
			n := Nested{Value: 1, Data: []byte{1, 2},
				Inner: &Nested{Value: 2, Data: []byte{3}}}
			c := Clone(n)
			assert.Equal(t, n.Value, c.Value)
			assert.Equal(t, n.Inner.Value, c.Inner.Value)
		}},
		{"cloneable", func() {
			s := stressCloneable{Value: 7, Data: []byte{0xFF}}
			c := Clone(s)
			assert.Equal(t, s.Value, c.Value)
		}},
		{"nil_slice", func() {
			var s []int
			c := Clone(s)
			assert.Nil(t, c)
		}},
		{"nil_map", func() {
			var m map[string]int
			c := Clone(m)
			assert.Nil(t, c)
		}},
		{"bool_slice", func() {
			s := []bool{true, false, true}
			c := Clone(s)
			assert.Equal(t, s, c)
		}},
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		for range goroutines {
			wg.Go(func() {
				for range iterations {
					src.fn()
				}
			})
		}
	}

	wg.Wait()
}

// TestConcurrentCloneIndependence verifies that clones produced
// concurrently are fully independent — mutations in one goroutine
// do not affect clones in another.
func TestConcurrentCloneIndependence(t *testing.T) {
	const goroutines = 100
	const iterations = 200

	original := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 6},
	}

	var wg sync.WaitGroup

	for i := range goroutines {
		wg.Go(func() {
			for j := range iterations {
				cloned := Clone(original)
				// Mutate the clone — must not affect original
				// or clones in other goroutines.
				cloned["a"][0] = i*1000 + j
				cloned["b"] = append(cloned["b"], i)
				// Original must remain unchanged.
				assert.Equal(t, 1, original["a"][0])
				assert.Len(t, original["b"], 3)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentClonePointerGraph stress-tests concurrent cloning
// of a shared pointer graph to verify pointer identity is preserved
// within each clone but independent across clones.
func TestConcurrentClonePointerGraph(t *testing.T) {
	const goroutines = 100
	const iterations = 200

	type Graph struct {
		A *int
		B *int // same pointer as A
	}

	shared := 99
	original := Graph{A: &shared, B: &shared}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				cloned := Clone(original)
				assert.Equal(t, 99, *cloned.A)
				assert.Equal(t, 99, *cloned.B)
				// Shared pointer identity preserved in clone.
				assert.Same(t, cloned.A, cloned.B)
				// Independent from original.
				assert.NotSame(t, &shared, cloned.A)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneWithCacheContention stress-tests the struct
// type cache under heavy contention by cloning many distinct struct
// types from many goroutines simultaneously.
func TestConcurrentCloneWithCacheContention(t *testing.T) {
	ResetCache()
	t.Cleanup(ResetCache)

	const goroutines = 200

	// Each goroutine clones a unique struct type plus shared types,
	// creating both cache misses and cache hits concurrently.
	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			// Shared types — cache hits after first population.
			for range 50 {
				cloneManyDistinctTypes()
			}
		})
	}

	wg.Wait()

	entries, fields := CacheStats()
	assert.Equal(t, 50, entries,
		"expected 50 cache entries, got %d", entries)
	assert.Greater(t, fields, 0)
}

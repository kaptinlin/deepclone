package deepclone

import (
	"bytes"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// stressCloner implements Cloner for concurrent testing.
type stressCloner struct {
	Value int
	Data  []byte
}

func (s stressCloner) Clone() (stressCloner, error) {
	return stressCloner{Value: s.Value, Data: bytes.Clone(s.Data)}, nil
}

// TestConcurrentCloneStructs stress-tests concurrent cloning of
// structs with slices and maps to verify data independence.
func TestConcurrentCloneStructs(t *testing.T) {
	t.Parallel()
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
				cloned := MustClone(original)
				if diff := cmp.Diff(original, cloned); diff != "" {
					t.Errorf("concurrent clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneSlices stress-tests concurrent cloning of
// typed slices through both fast paths and reflection paths.
func TestConcurrentCloneSlices(t *testing.T) {
	t.Parallel()
	const goroutines = 100
	const iterations = 500

	intSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	strSlice := []string{"a", "b", "c", "d", "e"}
	anySlice := []any{1, "two", 3.0, true, nil}

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				c := MustClone(intSlice)
				if diff := cmp.Diff(intSlice, c); diff != "" {
					t.Errorf("int slice clone mismatch (-want +got):\n%s", diff)
				}
				assert.Equal(t, cap(intSlice), cap(c))
			}
		})
		wg.Go(func() {
			for range iterations {
				c := MustClone(strSlice)
				if diff := cmp.Diff(strSlice, c); diff != "" {
					t.Errorf("string slice clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
		wg.Go(func() {
			for range iterations {
				c := MustClone(anySlice)
				if diff := cmp.Diff(anySlice, c); diff != "" {
					t.Errorf("any slice clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneMaps stress-tests concurrent cloning of
// typed maps through both fast paths and reflection paths.
func TestConcurrentCloneMaps(t *testing.T) {
	t.Parallel()
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
				c := MustClone(strMap)
				if diff := cmp.Diff(strMap, c); diff != "" {
					t.Errorf("string map clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
		wg.Go(func() {
			for range iterations {
				c := MustClone(intMap)
				if diff := cmp.Diff(intMap, c); diff != "" {
					t.Errorf("int map clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
		wg.Go(func() {
			for range iterations {
				c := MustClone(nestedMap)
				if diff := cmp.Diff(nestedMap, c); diff != "" {
					t.Errorf("nested map clone mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneCircularRef stress-tests concurrent cloning
// of structures with circular references to verify no infinite
// loops or data races occur.
func TestConcurrentCloneCircularRef(t *testing.T) {
	t.Parallel()
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
				cloned := MustClone(a)
				assert.Equal(t, a.ID, cloned.ID)
				assert.Equal(t, a.Name, cloned.Name)
				if diff := cmp.Diff(a.Tags, cloned.Tags); diff != "" {
					t.Errorf("circular tags mismatch (-want +got):\n%s", diff)
				}
				assert.Equal(t, b.ID, cloned.Next.ID)
				assert.Equal(t, c.ID, cloned.Next.Next.ID)
				// Verify circular ref is preserved.
				assert.Same(t, cloned, cloned.Next.Next.Next)
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneCloner stress-tests concurrent cloning of
// types implementing the Cloner interface.
func TestConcurrentCloneCloner(t *testing.T) {
	t.Parallel()
	const goroutines = 100
	const iterations = 500

	original := stressCloner{
		Value: 42,
		Data:  []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	cloned := MustClone(original)
	require.NotSame(t, &original.Data[0], &cloned.Data[0])

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for range iterations {
				cloned := MustClone(original)
				assert.Equal(t, original.Value, cloned.Value)
				if diff := cmp.Diff(original.Data, cloned.Data); diff != "" {
					t.Errorf("cloner data mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}

	wg.Wait()
}

// TestConcurrentCloneMixedTypes stress-tests concurrent cloning of
// many different types simultaneously to exercise all code paths
// (fast paths, Cloner, reflection) under contention.
func TestConcurrentCloneMixedTypes(t *testing.T) {
	t.Parallel()
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
			c := MustClone(12345)
			assert.Equal(t, 12345, c)
		}},
		{"string", func() {
			c := MustClone("concurrent")
			assert.Equal(t, "concurrent", c)
		}},
		{"pointer", func() {
			c := MustClone(&ptr)
			assert.Equal(t, ptr, *c)
			assert.NotSame(t, &ptr, c)
		}},
		{"int_slice", func() {
			s := []int{10, 20, 30}
			c := MustClone(s)
			if diff := cmp.Diff(s, c); diff != "" {
				t.Errorf("slice clone mismatch (-want +got):\n%s", diff)
			}
		}},
		{"string_map", func() {
			m := map[string]string{"k": "v"}
			c := MustClone(m)
			if diff := cmp.Diff(m, c); diff != "" {
				t.Errorf("map clone mismatch (-want +got):\n%s", diff)
			}
		}},
		{"nested_struct", func() {
			n := Nested{Value: 1, Data: []byte{1, 2},
				Inner: &Nested{Value: 2, Data: []byte{3}}}
			c := MustClone(n)
			assert.Equal(t, n.Value, c.Value)
			assert.Equal(t, n.Inner.Value, c.Inner.Value)
		}},
		{"cloner", func() {
			s := stressCloner{Value: 7, Data: []byte{0xFF}}
			c := MustClone(s)
			assert.Equal(t, s.Value, c.Value)
		}},
		{"nil_slice", func() {
			var s []int
			c := MustClone(s)
			assert.Nil(t, c)
		}},
		{"nil_map", func() {
			var m map[string]int
			c := MustClone(m)
			assert.Nil(t, c)
		}},
		{"bool_slice", func() {
			s := []bool{true, false, true}
			c := MustClone(s)
			if diff := cmp.Diff(s, c); diff != "" {
				t.Errorf("slice clone mismatch (-want +got):\n%s", diff)
			}
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
	t.Parallel()
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
				cloned := MustClone(original)
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
	t.Parallel()
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
				cloned := MustClone(original)
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
// This test runs serially because it asserts exact package-global cache counts.
func TestConcurrentCloneWithCacheContention(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

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

	entries, fields := cacheStats()
	assert.Equal(t, 50, entries,
		"expected 50 cache entries, got %d", entries)
	assert.Greater(t, fields, 0)
}

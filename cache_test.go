package deepclone

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run serially because they mutate or assert on the package-global struct cache.
func TestStructCacheMetrics(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	entries, fields := cacheStats()
	assert.Equal(t, 0, entries)
	assert.Equal(t, 0, fields)

	type TwoFields struct {
		A int
		B string
	}
	MustClone(TwoFields{A: 1, B: "x"})

	entries, fields = cacheStats()
	assert.Equal(t, 1, entries)
	assert.Equal(t, 2, fields)

	type ThreeFields struct {
		X float64
		Y float64
		Z float64
	}
	MustClone(ThreeFields{X: 1, Y: 2, Z: 3})

	entries, fields = cacheStats()
	assert.Equal(t, 2, entries)
	assert.Equal(t, 5, fields) // 2 + 3
}

func TestStructCacheMetricsIdempotent(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	type S struct{ V int }
	for range 100 {
		MustClone(S{V: 42})
	}

	entries, fields := cacheStats()
	assert.Equal(t, 1, entries, "same type cloned 100x produces one entry")
	assert.Equal(t, 1, fields)
}

func TestStructCacheMetricsIncludesUnexportedFields(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	type mixedVisibility struct {
		Visible string
		hidden  []int
		count   int
	}

	original := mixedVisibility{
		Visible: "x",
		count:   7,
	}
	cloned := MustClone(original)

	assert.Equal(t, "x", cloned.Visible)
	assert.Nil(t, cloned.hidden)
	assert.Equal(t, 7, cloned.count)

	entries, fields := cacheStats()
	assert.Equal(t, 1, entries)
	assert.Equal(t, 3, fields)
}

func TestStructCacheReset(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	type R struct{ V int }
	MustClone(R{V: 1})

	entries, _ := cacheStats()
	require.Equal(t, 1, entries)

	resetCache()
	entries, _ = cacheStats()
	assert.Equal(t, 0, entries)

	// Cache repopulates on next clone.
	MustClone(R{V: 2})
	entries, _ = cacheStats()
	assert.Equal(t, 1, entries)
}

// 50 distinct struct types used to populate the cache for memory tests.
// Each type is defined at package level so reflect.Type values are stable.
type cacheT01 struct{ F1 int }
type cacheT02 struct{ F1, F2 int }
type cacheT03 struct{ F1, F2, F3 int }
type cacheT04 struct{ F1, F2, F3, F4 int }
type cacheT05 struct{ F1, F2, F3, F4, F5 int }
type cacheT06 struct{ F1, F2, F3, F4, F5, F6 int }
type cacheT07 struct{ F1, F2, F3, F4, F5, F6, F7 int }
type cacheT08 struct{ F1, F2, F3, F4, F5, F6, F7, F8 int }
type cacheT09 struct{ F1, F2, F3, F4, F5, F6, F7, F8, F9 int }
type cacheT10 struct{ F1, F2, F3, F4, F5, F6, F7, F8, F9, F10 int }
type cacheT11 struct{ F1 string }
type cacheT12 struct{ F1, F2 string }
type cacheT13 struct{ F1, F2, F3 string }
type cacheT14 struct{ F1, F2, F3, F4 string }
type cacheT15 struct{ F1, F2, F3, F4, F5 string }
type cacheT16 struct{ F1 float64 }
type cacheT17 struct{ F1, F2 float64 }
type cacheT18 struct{ F1, F2, F3 float64 }
type cacheT19 struct{ F1, F2, F3, F4 float64 }
type cacheT20 struct{ F1, F2, F3, F4, F5 float64 }
type cacheT21 struct{ F1 bool }
type cacheT22 struct{ F1, F2 bool }
type cacheT23 struct{ F1, F2, F3 bool }
type cacheT24 struct{ F1, F2, F3, F4 bool }
type cacheT25 struct{ F1, F2, F3, F4, F5 bool }
type cacheT26 struct {
	A int
	B string
}
type cacheT27 struct {
	A int
	B string
	C float64
}
type cacheT28 struct {
	A int
	B string
	C float64
	D bool
}
type cacheT29 struct {
	A int
	B string
	C float64
	D bool
	E int
}
type cacheT30 struct {
	A int
	B string
	C float64
	D bool
	E int
	F string
}
type cacheT31 struct{ X []int }
type cacheT32 struct{ X []string }
type cacheT33 struct{ X map[string]int }
type cacheT34 struct{ X *int }
type cacheT35 struct{ X *string }
type cacheT36 struct {
	X []int
	Y string
}
type cacheT37 struct {
	X []string
	Y int
}
type cacheT38 struct {
	X map[string]int
	Y bool
}
type cacheT39 struct {
	X *int
	Y float64
}
type cacheT40 struct {
	X *string
	Y int
}
type cacheT41 struct{ A, B, C, D, E, F, G, H, I, J int }
type cacheT42 struct{ A, B, C, D, E, F, G, H, I, J string }
type cacheT43 struct{ A, B, C, D, E, F, G, H, I, J float64 }
type cacheT44 struct{ A, B, C, D, E, F, G, H, I, J bool }
type cacheT45 struct {
	A int
	B []int
	C map[string]int
	D *int
	E string
}
type cacheT46 struct {
	A string
	B []string
	C map[string]string
	D *string
	E int
}
type cacheT47 struct {
	A float64
	B []float64
	C map[string]float64
	D *float64
	E bool
}
type cacheT48 struct{ A, B, C, D, E, F, G, H, I, J, K, L, M, N, O int }
type cacheT49 struct{ A, B, C, D, E, F, G, H, I, J, K, L, M, N, O string }
type cacheT50 struct{ A, B, C, D, E, F, G, H, I, J, K, L, M, N, O float64 }

// cloneManyDistinctTypes populates the cache with 50 distinct struct types.
func cloneManyDistinctTypes() {
	MustClone(cacheT01{})
	MustClone(cacheT02{})
	MustClone(cacheT03{})
	MustClone(cacheT04{})
	MustClone(cacheT05{})
	MustClone(cacheT06{})
	MustClone(cacheT07{})
	MustClone(cacheT08{})
	MustClone(cacheT09{})
	MustClone(cacheT10{})
	MustClone(cacheT11{})
	MustClone(cacheT12{})
	MustClone(cacheT13{})
	MustClone(cacheT14{})
	MustClone(cacheT15{})
	MustClone(cacheT16{})
	MustClone(cacheT17{})
	MustClone(cacheT18{})
	MustClone(cacheT19{})
	MustClone(cacheT20{})
	MustClone(cacheT21{})
	MustClone(cacheT22{})
	MustClone(cacheT23{})
	MustClone(cacheT24{})
	MustClone(cacheT25{})
	MustClone(cacheT26{})
	MustClone(cacheT27{})
	MustClone(cacheT28{})
	MustClone(cacheT29{})
	MustClone(cacheT30{})
	MustClone(cacheT31{})
	MustClone(cacheT32{})
	MustClone(cacheT33{})
	MustClone(cacheT34{})
	MustClone(cacheT35{})
	MustClone(cacheT36{})
	MustClone(cacheT37{})
	MustClone(cacheT38{})
	MustClone(cacheT39{})
	MustClone(cacheT40{})
	MustClone(cacheT41{})
	MustClone(cacheT42{})
	MustClone(cacheT43{})
	MustClone(cacheT44{})
	MustClone(cacheT45{})
	MustClone(cacheT46{})
	MustClone(cacheT47{})
	MustClone(cacheT48{})
	MustClone(cacheT49{})
	MustClone(cacheT50{})
}

// TestCacheMemoryFootprint validates that the struct field cache uses
// bounded, predictable memory. This demonstrates why LRU eviction is
// unnecessary: entries equal distinct struct types — a finite quantity.
func TestCacheMemoryFootprint(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	// Use TotalAlloc (monotonically increasing) to measure cumulative
	// allocations. HeapAlloc can decrease due to GC between measurements.
	var before runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	// Clone 50 distinct struct types with varying field counts.
	cloneManyDistinctTypes()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	entries, fields := cacheStats()
	assert.Equal(t, 50, entries)
	assert.Greater(t, fields, 0)

	// Each structTypeInfo stores a small field metadata slice plus map bucket
	// overhead per entry.
	//
	// For 50 types averaging ~5 fields, total should be well under 1 MB.
	totalAlloc := after.TotalAlloc - before.TotalAlloc
	const maxExpected = 1 << 20 // 1 MB
	assert.Less(t, totalAlloc, uint64(maxExpected),
		"cache for 50 types should use well under 1 MB; got %d bytes",
		totalAlloc)

	t.Logf("cache: %d entries, %d fields, total alloc: %d bytes",
		entries, fields, totalAlloc)
	if entries > 0 {
		t.Logf("estimated per-entry cost: %d bytes",
			totalAlloc/uint64(entries))
	}
}

// TestCacheBoundedGrowth verifies that cloning the same types repeatedly
// does not grow the cache beyond the number of distinct types.
func TestCacheBoundedGrowth(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	// Populate with 50 types.
	cloneManyDistinctTypes()
	entries1, fields1 := cacheStats()

	// Clone all 50 types again 100 times.
	for range 100 {
		cloneManyDistinctTypes()
	}

	entries2, fields2 := cacheStats()
	assert.Equal(t, entries1, entries2, "cache entries should not grow")
	assert.Equal(t, fields1, fields2, "cached fields should not grow")
}

// TestCacheConcurrentAccess verifies thread safety of the cache under
// concurrent clone operations from multiple goroutines.
func TestCacheConcurrentAccess(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	const goroutines = 50
	var wg sync.WaitGroup

	// Each goroutine clones all 50 types concurrently.
	for range goroutines {
		wg.Go(func() {
			cloneManyDistinctTypes()
		})
	}
	wg.Wait()

	entries, _ := cacheStats()
	assert.Equal(t, 50, entries,
		"concurrent access should produce exactly 50 entries")
}

// TestStructCacheResetConcurrent verifies that struct cache reset is safe to call
// concurrently with clone operations.
func TestStructCacheResetConcurrent(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	const goroutines = 20
	var wg sync.WaitGroup

	for i := range goroutines {
		wg.Go(func() {
			cloneManyDistinctTypes()
			if i%5 == 0 {
				resetCache()
			}
		})
	}
	wg.Wait()

	// After all goroutines finish, cache should be in a valid state.
	// The exact count depends on timing, but it must not panic.
	entries, fields := cacheStats()
	assert.GreaterOrEqual(t, entries, 0)
	assert.GreaterOrEqual(t, fields, 0)
}

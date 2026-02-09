package deepclone

import "testing"

// Benchmark data types.
type benchSimple struct {
	ID   int
	Name string
	Age  int
}

type benchUserSettings struct {
	Theme         string
	Language      string
	Notifications bool
}

type benchProfile struct {
	Email    string
	Settings *benchUserSettings
}

type benchNested struct {
	ID       int
	Name     string
	Profile  *benchProfile
	Tags     []string
	Settings map[string]any
}

type benchCircular struct {
	ID   int
	Name string
	Self *benchCircular
}

// Benchmark fixtures.
var (
	benchIntVal    = 42
	benchStringVal = "hello world"
	benchSliceVal  = func() []int {
		s := make([]int, 100)
		for i := range s {
			s[i] = i
		}
		return s
	}()
	benchMapVal = func() map[string]int {
		m := make(map[string]int, 100)
		for i := range 100 {
			k := string(rune('a'+i%26)) + string(rune('a'+(i/26)%26))
			m[k] = i
		}
		return m
	}()
	benchSimpleVal = benchSimple{ID: 1, Name: "test", Age: 25}
	benchNestedVal = benchNested{
		ID:   1,
		Name: "test",
		Profile: &benchProfile{
			Email: "test@example.com",
			Settings: &benchUserSettings{
				Theme:         "dark",
				Language:      "en",
				Notifications: true,
			},
		},
		Tags:     []string{"tag1", "tag2", "tag3"},
		Settings: map[string]any{"key1": "value1", "key2": 42},
	}
	benchCircularVal = func() *benchCircular {
		c := &benchCircular{ID: 1, Name: "circular"}
		c.Self = c
		return c
	}()
	benchLargeSliceVal = func() []int {
		s := make([]int, 10000)
		for i := range s {
			s[i] = i
		}
		return s
	}()
)

// BenchmarkClone uses b.Loop() (Go 1.24+) for all sub-benchmarks.
func BenchmarkClone(b *testing.B) {
	b.Run("int", func(b *testing.B) {
		for b.Loop() {
			_ = Clone(benchIntVal)
		}
	})

	b.Run("string", func(b *testing.B) {
		for b.Loop() {
			_ = Clone(benchStringVal)
		}
	})

	b.Run("slice_100", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchSliceVal)
		}
	})

	b.Run("map_100", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchMapVal)
		}
	})

	b.Run("simple_struct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchSimpleVal)
		}
	})

	b.Run("nested_struct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchNestedVal)
		}
	})

	b.Run("pointer", func(b *testing.B) {
		ptr := &benchSimpleVal
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(ptr)
		}
	})

	b.Run("circular", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchCircularVal)
		}
	})

	b.Run("large_slice_10k", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(benchLargeSliceVal)
		}
	})

	b.Run("interface", func(b *testing.B) {
		var iface any = benchSimpleVal
		b.ReportAllocs()
		for b.Loop() {
			_ = Clone(iface)
		}
	})
}

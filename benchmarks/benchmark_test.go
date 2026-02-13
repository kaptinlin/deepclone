package benchmarks

import (
	"github.com/go-json-experiment/json"
	"reflect"
	"testing"

	"github.com/kaptinlin/deepclone"

	// Competing libraries
	"github.com/huandu/go-clone"
	"github.com/jinzhu/copier"
	"github.com/mohae/deepcopy"
	reflectpkg "golang.design/x/reflect"
)

// Test data structures
type SimpleStruct struct {
	ID   int
	Name string
	Age  int
}

type NestedStruct struct {
	ID       int
	Name     string
	Profile  *Profile
	Tags     []string
	Settings map[string]any
}

type Profile struct {
	Email    string
	Settings *UserSettings
}

type UserSettings struct {
	Theme         string
	Language      string
	Notifications bool
}

type CircularStruct struct {
	ID   int
	Name string
	Self *CircularStruct
}

// Test data
var (
	testInt    = 42
	testString = "hello world"
	testFloat  = 3.14159
	testBool   = true
	testSlice  = make([]int, 100)
	testMap    = make(map[string]int, 100)
	testSimple = SimpleStruct{ID: 1, Name: "test", Age: 25}
	testNested = NestedStruct{
		ID:   1,
		Name: "test",
		Profile: &Profile{
			Email: "test@example.com",
			Settings: &UserSettings{
				Theme:         "dark",
				Language:      "en",
				Notifications: true,
			},
		},
		Tags:     []string{"tag1", "tag2", "tag3"},
		Settings: map[string]any{"key1": "value1", "key2": 42},
	}
	testCircular = func() *CircularStruct {
		c := &CircularStruct{ID: 1, Name: "circular"}
		c.Self = c
		return c
	}()
)

func init() {
	// Initialize test data
	for i := range 100 {
		testSlice[i] = i
		testMap[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = i
	}
}

// Benchmark: Int
func BenchmarkInt_This(b *testing.B) {
	for b.Loop() {
		_ = deepclone.Clone(testInt)
	}
}

func BenchmarkInt_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testInt)
	}
}

func BenchmarkInt_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testInt)
	}
}

func BenchmarkInt_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testInt)
	}
}

// Benchmark: String
func BenchmarkString_This(b *testing.B) {
	for b.Loop() {
		_ = deepclone.Clone(testString)
	}
}

func BenchmarkString_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testString)
	}
}

func BenchmarkString_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testString)
	}
}

func BenchmarkString_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testString)
	}
}

// Benchmark: Float64
func BenchmarkFloat_This(b *testing.B) {
	for b.Loop() {
		_ = deepclone.Clone(testFloat)
	}
}

func BenchmarkFloat_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testFloat)
	}
}

func BenchmarkFloat_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testFloat)
	}
}

func BenchmarkFloat_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testFloat)
	}
}

// Benchmark: Bool
func BenchmarkBool_This(b *testing.B) {
	for b.Loop() {
		_ = deepclone.Clone(testBool)
	}
}

func BenchmarkBool_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testBool)
	}
}

func BenchmarkBool_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testBool)
	}
}

func BenchmarkBool_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testBool)
	}
}

// Benchmark: Slice (100 ints)
func BenchmarkSlice100_This(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = deepclone.Clone(testSlice)
	}
}

func BenchmarkSlice100_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testSlice)
	}
}

func BenchmarkSlice100_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testSlice)
	}
}

func BenchmarkSlice100_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testSlice)
	}
}

// Benchmark: Map (100 entries)
func BenchmarkMap100_This(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = deepclone.Clone(testMap)
	}
}

func BenchmarkMap100_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testMap)
	}
}

func BenchmarkMap100_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testMap)
	}
}

func BenchmarkMap100_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testMap)
	}
}

// Benchmark: Simple Struct
func BenchmarkSimpleStruct_This(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = deepclone.Clone(testSimple)
	}
}

func BenchmarkSimpleStruct_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testSimple)
	}
}

func BenchmarkSimpleStruct_Jinzhu(b *testing.B) {
	for b.Loop() {
		var dst SimpleStruct
		_ = copier.Copy(&dst, &testSimple)
	}
}

func BenchmarkSimpleStruct_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testSimple)
	}
}

func BenchmarkSimpleStruct_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testSimple)
	}
}

// Benchmark: Nested Struct
func BenchmarkNestedStruct_This(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = deepclone.Clone(testNested)
	}
}

func BenchmarkNestedStruct_Mohae(b *testing.B) {
	for b.Loop() {
		_ = deepcopy.Copy(testNested)
	}
}

func BenchmarkNestedStruct_Jinzhu(b *testing.B) {
	for b.Loop() {
		var dst NestedStruct
		_ = copier.CopyWithOption(&dst, &testNested, copier.Option{DeepCopy: true})
	}
}

func BenchmarkNestedStruct_Huandu(b *testing.B) {
	for b.Loop() {
		_ = clone.Clone(testNested)
	}
}

func BenchmarkNestedStruct_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testNested)
	}
}

// Benchmark: Pointers
func BenchmarkPointer_This(b *testing.B) {
	ptr := &testSimple
	for b.Loop() {
		_ = deepclone.Clone(ptr)
	}
}

func BenchmarkPointer_Mohae(b *testing.B) {
	ptr := &testSimple
	for b.Loop() {
		_ = deepcopy.Copy(ptr)
	}
}

func BenchmarkPointer_Huandu(b *testing.B) {
	ptr := &testSimple
	for b.Loop() {
		_ = clone.Clone(ptr)
	}
}

func BenchmarkPointer_GolangDesign(b *testing.B) {
	ptr := &testSimple
	for b.Loop() {
		_ = reflectpkg.DeepCopy(ptr)
	}
}

// Benchmark: Circular Reference
func BenchmarkCircular_This(b *testing.B) {
	for b.Loop() {
		_ = deepclone.Clone(testCircular)
	}
}

// mohae/deepcopy doesn't handle circular references properly - causes stack overflow
// func BenchmarkCircular_Mohae(b *testing.B) {
// 	for b.Loop() {
// 		_ = deepcopy.Copy(testCircular)
// 	}
// }

// huandu/go-clone also has issues with circular references
// func BenchmarkCircular_Huandu(b *testing.B) {
// 	for b.Loop() {
// 		_ = clone.Clone(testCircular)
// 	}
// }

func BenchmarkCircular_GolangDesign(b *testing.B) {
	for b.Loop() {
		_ = reflectpkg.DeepCopy(testCircular)
	}
}

// Benchmark: Interface
func BenchmarkInterface_This(b *testing.B) {
	var iface any = testSimple
	for b.Loop() {
		_ = deepclone.Clone(iface)
	}
}

func BenchmarkInterface_Mohae(b *testing.B) {
	var iface any = testSimple
	for b.Loop() {
		_ = deepcopy.Copy(iface)
	}
}

func BenchmarkInterface_Huandu(b *testing.B) {
	var iface any = testSimple
	for b.Loop() {
		_ = clone.Clone(iface)
	}
}

func BenchmarkInterface_GolangDesign(b *testing.B) {
	var iface any = testSimple
	for b.Loop() {
		_ = reflectpkg.DeepCopy(iface)
	}
}

// Large data benchmarks
func BenchmarkLargeSlice_This(b *testing.B) {
	largeSlice := make([]int, 10000)
	for i := range largeSlice {
		largeSlice[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = deepclone.Clone(largeSlice)
	}
}

func BenchmarkLargeSlice_Mohae(b *testing.B) {
	largeSlice := make([]int, 10000)
	for i := range largeSlice {
		largeSlice[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		_ = deepcopy.Copy(largeSlice)
	}
}

func BenchmarkLargeSlice_Huandu(b *testing.B) {
	largeSlice := make([]int, 10000)
	for i := range largeSlice {
		largeSlice[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		_ = clone.Clone(largeSlice)
	}
}

func BenchmarkLargeSlice_GolangDesign(b *testing.B) {
	largeSlice := make([]int, 10000)
	for i := range largeSlice {
		largeSlice[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		_ = reflectpkg.DeepCopy(largeSlice)
	}
}

// JSON vs Deep Clone comparison
func BenchmarkJSON_Marshal_Unmarshal(b *testing.B) {
	for b.Loop() {
		data, _ := json.Marshal(testNested)
		var dst NestedStruct
		_ = json.Unmarshal(data, &dst)
	}
}

// Reflection-based manual clone
func BenchmarkReflection_Manual(b *testing.B) {
	for b.Loop() {
		_ = manualReflectClone(testNested)
	}
}

func manualReflectClone(src any) any {
	srcVal := reflect.ValueOf(src)
	return reflect.New(srcVal.Type()).Elem().Interface()
}

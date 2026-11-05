package gosimruntime

import (
	"bytes"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

func TestHashBytes(t *testing.T) {
	for n := 0; n < 257; n++ {
		data := bytes.Repeat([]byte{byte(n)}, n)
		got := HashBytes(data, 123)
		if got != HashBytes(bytes.Clone(data), 123) {
			t.Fatalf("length %d depends on allocation", n)
		}
		if got == HashBytes(data, 456) {
			t.Fatalf("length %d ignores seed", n)
		}
	}
}

func TestHashComparableEquality(t *testing.T) {
	pairs := [][2]any{
		{float64(0), math.Copysign(0, -1)},
		{complex(0.0, math.Copysign(0, -1)), complex(math.Copysign(0, -1), 0.0)},
		{[2]string{"a", "bc"}, [2]string{"a", "bc"}},
		{struct {
			private int
			Value   any
		}{7, "value"}, struct {
			private int
			Value   any
		}{7, "value"}},
		{nil, nil},
	}
	for _, pair := range pairs {
		if HashComparable(pair[0], 99) != HashComparable(pair[1], 99) {
			t.Errorf("equal values hash differently: %#v", pair)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("incomparable value did not panic")
		}
	}()
	HashComparable(struct{ _ []int }{}, 99)
}

func TestHashComparablePointerRunIdentity(t *testing.T) {
	var first []uint64
	for attempt := range 2 {
		var got []uint64
		run(func() {
			a, b := new(int), new(int)
			for _, value := range []any{a, b, a, (*int)(nil), make(chan int), math.NaN(), math.NaN()} {
				got = append(got, HashComparable(value, 123))
			}
		}, 1, false, false, "INFO", makeConsoleLogger(os.Stderr), nil)
		if got[0] != got[2] || got[0] == got[1] || got[1] == got[3] {
			t.Fatal("pointer identity not preserved", got)
		}
		if attempt == 0 {
			first = got
		} else if !reflect.DeepEqual(first, got) {
			t.Fatalf("run state leaked: %v != %v", first, got)
		}
	}
}

// Reflect must handle non-exported members without Interface(), preserve Go's
// equality rules for blank fields, and reject dynamically unhashable interfaces.
func TestHashComparableReflectEdges(t *testing.T) {
	type padded struct {
		B byte
		_ uint64
		N int
	}
	a, b := padded{B: 1, N: 4}, padded{B: 1, N: 4}
	// Modify blank storage and padding, neither of which participates in equality.
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&b)), unsafe.Sizeof(b))
	for i := 1; i < int(unsafe.Offsetof(b.N)); i++ {
		raw[i] = 0xff
	}
	if a != b || HashComparable(a, 7) != HashComparable(b, 7) {
		t.Fatal("blank fields or padding entered hash")
	}
	type private struct {
		p     *int
		ch    chan int
		ptr   unsafe.Pointer
		value any
	}
	var hashes [2]uint64
	run(func() {
		p := new(int)
		v := private{p, make(chan int), unsafe.Pointer(p), (*int)(nil)}
		hashes[0], hashes[1] = HashComparable(v, 7), HashComparable(v, 7)
	}, 1, false, false, "INFO", makeConsoleLogger(os.Stderr), nil)
	if hashes[0] != hashes[1] {
		t.Fatal("private fields changed hash")
	}
	for _, value := range []any{[]int(nil), map[int]int(nil), (func())(nil), struct{ V any }{[]int{1}}, [1]any{map[int]int{}}} {
		t.Run(reflect.TypeOf(value).String(), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("incomparable value accepted")
				}
			}()
			HashComparable(value, 7)
		})
	}
}

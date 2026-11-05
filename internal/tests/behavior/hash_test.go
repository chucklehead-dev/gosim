//go:build sim

package behavior_test

import (
	"bytes"
	"fmt"
	"hash/maphash"
	"log/slog"
	"math"
	"testing"
)

func TestRandMapHashCoverage(t *testing.T) {
	seed := maphash.MakeSeed()
	var results []uint64
	for _, n := range []int{0, 1, 3, 4, 8, 16, 17, 48, 49, 256, 257, 1024} {
		data := bytes.Repeat([]byte{'x'}, n)
		want := maphash.Bytes(seed, data)
		if want != maphash.String(seed, string(data)) {
			t.Fatalf("Bytes/String disagree at length %d", n)
		}
		var h maphash.Hash
		h.SetSeed(seed)
		for i := range data {
			h.Write(data[i : i+1])
		}
		if h.Sum64() != want {
			t.Fatalf("segmented hash disagrees at length %d", n)
		}
		results = append(results, want)
	}
	slog.Info("rand", "val", fmt.Sprint(results))
}

func TestRandMapHashComparable(t *testing.T) {
	seed := maphash.MakeSeed()
	a, b := new(int), new(int)
	var results []uint64
	values := []any{nil, true, int64(-4), uint64(7), "abc", [2]string{"a", "bc"}, struct {
		private string
		N       int
	}{"x", 2}, a, b, a, (*int)(nil), make(chan int), math.NaN()}
	for _, v := range values {
		results = append(results, maphash.Comparable(seed, v))
	}
	if results[7] != results[9] || results[7] == results[8] {
		t.Fatal("pointer identity lost")
	}
	if maphash.Comparable(seed, 0.0) != maphash.Comparable(seed, math.Copysign(0, -1)) {
		t.Fatal("equal signed zero hashes differ")
	}
	var h maphash.Hash
	h.SetSeed(seed)
	maphash.WriteComparable(&h, struct {
		Key string
		N   int
	}{"entry", 42})
	results = append(results, h.Sum64())
	slog.Info("rand", "val", fmt.Sprint(results))
}

func TestMapHashRejectsMaps(t *testing.T) {
	type named map[int]int
	for _, value := range []any{
		map[int]int{},
		(map[int]int)(nil),
		named{},
		[1]map[int]int{},
		struct{ Value any }{map[int]int{}},
		struct{ _ map[int]int }{},
	} {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("map value accepted as comparable")
				}
			}()
			maphash.Comparable(maphash.MakeSeed(), value)
		})
	}
	// Pointer targets do not participate in comparability.
	m := map[int]int{}
	seed := maphash.MakeSeed()
	if maphash.Comparable(seed, &m) != maphash.Comparable(seed, &m) {
		t.Fatal("pointer to map lost identity")
	}
}

func TestMapHashWriteComparableRejectsMap(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("WriteComparable accepted map in interface")
		}
	}()
	var h maphash.Hash
	h.SetSeed(maphash.MakeSeed())
	maphash.WriteComparable[any](&h, map[int]int{})
}

// Copyright 2026 The gosim Authors.
// Byte mixing is adapted from src/hash/maphash/maphash_purego.go in Go 1.26.10:
// https://cs.opensource.google/go/go/+/refs/tags/go1.26.10:src/hash/maphash/maphash_purego.go
// Copyright 2023 The Go Authors. All rights reserved.
// Use of the adapted code is governed by the BSD-style license in LICENSE_GO.

package gosimruntime

import (
	"encoding/binary"
	"math"
	"math/bits"
	"reflect"
	"unsafe"
)

// HashBytes is the simulation's seeded, process-independent byte hash. A seed
// comes from simulation entropy; the mixer never uses the host runtime's secret
// hash keys. Like hash/maphash, this is not a cryptographic hash.
func HashBytes(data []byte, seed uint64) uint64 {
	if len(data) == 0 {
		return seed
	}
	const k0, k1, k2, k3 = uint64(0xa0761d6478bd642f), uint64(0xe7037ed1a0b428db), uint64(0x8ebc6af09c88c6e3), uint64(0x589965cc75374cc3)
	mix := func(a, b uint64) uint64 { hi, lo := bits.Mul64(a, b); return hi ^ lo }
	read := binary.LittleEndian.Uint64
	p, n := data, len(data)
	seed ^= k0
	if n > 48 {
		s1, s2 := seed, seed
		for n > 48 {
			seed = mix(read(p)^k1, read(p[8:])^seed)
			s1 = mix(read(p[16:])^k2, read(p[24:])^s1)
			s2 = mix(read(p[32:])^k3, read(p[40:])^s2)
			p, n = p[48:], n-48
		}
		seed ^= s1 ^ s2
	}
	for n > 16 {
		seed = mix(read(p)^k1, read(p[8:])^seed)
		p, n = p[16:], n-16
	}
	var a, b uint64
	if n < 4 {
		a = uint64(p[0])<<16 | uint64(p[n>>1])<<8 | uint64(p[n-1])
	} else {
		j := (n >> 3) << 2
		a = uint64(binary.LittleEndian.Uint32(p))<<32 | uint64(binary.LittleEndian.Uint32(p[j:]))
		b = uint64(binary.LittleEndian.Uint32(p[n-4:]))<<32 | uint64(binary.LittleEndian.Uint32(p[n-4-j:]))
	}
	return mix(0x1d8e4e27c47d124f^uint64(len(data)), mix(a^k1, b^seed))
}

// HashComparable hashes Go equality values without ABI-generated host hashers.
// Pointer identity is represented by an ID assigned on first observation in the
// run; neither addresses, padding, nor host-process hash seeds enter the bytes.
func HashComparable(value any, seed uint64) uint64 {
	return HashBytes(appendHashValue(nil, reflect.ValueOf(value)), seed)
}

func appendHashValue(out []byte, value reflect.Value) []byte {
	if !value.IsValid() {
		return append(out, 0)
	}
	if !value.Type().Comparable() {
		panic("maphash: hash of unhashable type " + value.Type().String())
	}
	out = append(out, byte(value.Kind()))
	name := value.Type().PkgPath() + "/" + value.Type().String()
	out = binary.LittleEndian.AppendUint64(out, uint64(len(name)))
	out = append(out, name...)
	switch value.Kind() {
	case reflect.Bool:
		if value.Bool() {
			return append(out, 1)
		}
		return append(out, 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return binary.LittleEndian.AppendUint64(out, uint64(value.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return binary.LittleEndian.AppendUint64(out, value.Uint())
	case reflect.Float32, reflect.Float64:
		return appendHashFloat(out, value.Float())
	case reflect.Complex64, reflect.Complex128:
		v := value.Complex()
		return appendHashFloat(appendHashFloat(out, real(v)), imag(v))
	case reflect.String:
		v := value.String()
		out = binary.LittleEndian.AppendUint64(out, uint64(len(v)))
		return append(out, v...)
	case reflect.Array:
		for i := range value.Len() {
			out = appendHashValue(out, value.Index(i))
		}
		return out
	case reflect.Struct:
		for i := range value.NumField() {
			if value.Type().Field(i).Name != "_" { // blank fields do not participate in ==
				out = appendHashValue(out, value.Field(i))
			}
		}
		return out
	case reflect.Interface:
		return appendHashValue(out, value.Elem())
	case reflect.Pointer, reflect.UnsafePointer, reflect.Chan:
		return binary.LittleEndian.AppendUint64(out, hashPointerID(value.UnsafePointer()))
	default:
		panic("maphash: hash of unhashable type " + value.Type().String())
	}
}

func appendHashFloat(out []byte, value float64) []byte {
	var bits uint64
	switch {
	case value == 0: // +0 == -0, including each component of a complex value.
	case math.IsNaN(value):
		bits = Fastrand64() // NaN is unequal to itself; vary each hash deterministically.
	default:
		bits = math.Float64bits(value)
	}
	return binary.LittleEndian.AppendUint64(out, bits)
}

//go:norace
func hashPointerID(pointer unsafe.Pointer) uint64 {
	if pointer == nil {
		return 0
	}
	s := gs.get()
	if s.hashPointers == nil {
		s.hashPointers = make(map[unsafe.Pointer]uint64)
	}
	if id := s.hashPointers[pointer]; id != 0 {
		return id
	}
	id := uint64(len(s.hashPointers)) + 1
	// An unsafe.Pointer key is a GC root, unlike uintptr. Keep the object alive
	// until this run ends so address reuse cannot change its identity.
	s.hashPointers[pointer] = id
	return id
}

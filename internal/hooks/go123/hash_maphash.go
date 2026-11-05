package go123

import (
	"unsafe"

	"github.com/chucklehead-dev/gosim/gosimruntime"
)

func HashMaphash_runtime_rand() uint64 {
	return gosimruntime.Fastrand64()
}

// HashMaphash_runtime_memhash replaces the host's process-keyed hash.
func HashMaphash_runtime_memhash(p unsafe.Pointer, seed, size uintptr) uintptr {
	return uintptr(gosimruntime.HashBytes(unsafe.Slice((*byte)(p), size), uint64(seed)))
}

func HashMaphash_Comparable(value any, seed uint64) uint64 {
	return gosimruntime.HashComparable(value, seed)
}

// Native map machinery is not translated. maphash only needs this target-width constant.
const InternalRuntimeMaps_Use64BitHash = unsafe.Sizeof(uintptr(0)) == 8

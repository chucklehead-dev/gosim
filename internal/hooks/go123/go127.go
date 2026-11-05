package go123

import (
	"unsafe"

	"github.com/chucklehead-dev/gosim/internal/race"
)

// Go 1.27's sync.Pool calls internal/runtime/atomic directly. Keep those
// operations at the same cooperative atomic boundary as sync/atomic instead of
// retaining host assembly that bypasses the simulation.

func runtimeAtomicByte() {
	if race.Enabled {
		panic("gosim: internal/runtime/atomic byte operations with -race are unsupported")
	}
	maybeAtomicYield()
}

// The FIPS CPU-jitter entropy source cannot run against simulated monotonic
// time: busy-loop readings do not advance virtual time. Refuse it explicitly
// rather than introduce host timing or claim a simulated FIPS entropy source.
func CryptoInternalFips140depsTime_monoTime() int64 {
	panic("gosim: FIPS CPU-jitter entropy timing is unsupported; use ordinary crypto/rand for deterministic simulation")
}

func InternalRuntimeAtomic_Load(ptr *uint32) uint32 {
	return SyncAtomic_LoadUint32(ptr)
}

func InternalRuntimeAtomic_Load64(ptr *uint64) uint64 {
	return SyncAtomic_LoadUint64(ptr)
}

func InternalRuntimeAtomic_LoadAcq(ptr *uint32) uint32 {
	return SyncAtomic_LoadUint32(ptr)
}

func InternalRuntimeAtomic_LoadAcq64(ptr *uint64) uint64 {
	return SyncAtomic_LoadUint64(ptr)
}

func InternalRuntimeAtomic_LoadAcquintptr(ptr *uintptr) uintptr {
	return SyncAtomic_LoadUintptr(ptr)
}

func InternalRuntimeAtomic_Loaduintptr(ptr *uintptr) uintptr {
	return SyncAtomic_LoadUintptr(ptr)
}

func InternalRuntimeAtomic_Loaduint(ptr *uint) uint {
	return uint(SyncAtomic_LoadUintptr((*uintptr)(unsafe.Pointer(ptr))))
}

func InternalRuntimeAtomic_Loadint32(ptr *int32) int32 {
	return SyncAtomic_LoadInt32(ptr)
}

func InternalRuntimeAtomic_Loadint64(ptr *int64) int64 {
	return SyncAtomic_LoadInt64(ptr)
}

func InternalRuntimeAtomic_Loadp(ptr unsafe.Pointer) unsafe.Pointer {
	return SyncAtomic_LoadPointer((*unsafe.Pointer)(ptr))
}

func InternalRuntimeAtomic_Cas(ptr *uint32, old, next uint32) bool {
	return SyncAtomic_CompareAndSwapUint32(ptr, old, next)
}

func InternalRuntimeAtomic_Cas64(ptr *uint64, old, next uint64) bool {
	return SyncAtomic_CompareAndSwapUint64(ptr, old, next)
}

func InternalRuntimeAtomic_CasRel(ptr *uint32, old, next uint32) bool {
	return SyncAtomic_CompareAndSwapUint32(ptr, old, next)
}

func InternalRuntimeAtomic_Casp1(ptr *unsafe.Pointer, old, next unsafe.Pointer) bool {
	return SyncAtomic_CompareAndSwapPointer(ptr, old, next)
}

func InternalRuntimeAtomic_Casint32(ptr *int32, old, next int32) bool {
	return SyncAtomic_CompareAndSwapInt32(ptr, old, next)
}

func InternalRuntimeAtomic_Casint64(ptr *int64, old, next int64) bool {
	return SyncAtomic_CompareAndSwapInt64(ptr, old, next)
}

func InternalRuntimeAtomic_Casuintptr(ptr *uintptr, old, next uintptr) bool {
	return SyncAtomic_CompareAndSwapUintptr(ptr, old, next)
}

func InternalRuntimeAtomic_Xadd(ptr *uint32, delta int32) uint32 {
	return SyncAtomic_AddUint32(ptr, uint32(delta))
}

func InternalRuntimeAtomic_Xadd64(ptr *uint64, delta int64) uint64 {
	return SyncAtomic_AddUint64(ptr, uint64(delta))
}

func InternalRuntimeAtomic_Xadduintptr(ptr *uintptr, delta uintptr) uintptr {
	return SyncAtomic_AddUintptr(ptr, delta)
}

func InternalRuntimeAtomic_Xaddint32(ptr *int32, delta int32) int32 {
	return SyncAtomic_AddInt32(ptr, delta)
}

func InternalRuntimeAtomic_Xaddint64(ptr *int64, delta int64) int64 {
	return SyncAtomic_AddInt64(ptr, delta)
}

func InternalRuntimeAtomic_Xchg(ptr *uint32, next uint32) uint32 {
	return SyncAtomic_SwapUint32(ptr, next)
}

func InternalRuntimeAtomic_Xchg64(ptr *uint64, next uint64) uint64 {
	return SyncAtomic_SwapUint64(ptr, next)
}

func InternalRuntimeAtomic_Xchguintptr(ptr *uintptr, next uintptr) uintptr {
	return SyncAtomic_SwapUintptr(ptr, next)
}

func InternalRuntimeAtomic_Xchgint32(ptr *int32, next int32) int32 {
	return SyncAtomic_SwapInt32(ptr, next)
}

func InternalRuntimeAtomic_Xchgint64(ptr *int64, next int64) int64 {
	return SyncAtomic_SwapInt64(ptr, next)
}

func InternalRuntimeAtomic_Store(ptr *uint32, value uint32) {
	SyncAtomic_StoreUint32(ptr, value)
}

func InternalRuntimeAtomic_Store64(ptr *uint64, value uint64) {
	SyncAtomic_StoreUint64(ptr, value)
}

func InternalRuntimeAtomic_StoreRel(ptr *uint32, value uint32) {
	SyncAtomic_StoreUint32(ptr, value)
}

func InternalRuntimeAtomic_StoreRel64(ptr *uint64, value uint64) {
	SyncAtomic_StoreUint64(ptr, value)
}

func InternalRuntimeAtomic_StoreReluintptr(ptr *uintptr, value uintptr) {
	SyncAtomic_StoreUintptr(ptr, value)
}

func InternalRuntimeAtomic_Storeuintptr(ptr *uintptr, value uintptr) {
	SyncAtomic_StoreUintptr(ptr, value)
}

func InternalRuntimeAtomic_Storeint32(ptr *int32, value int32) {
	SyncAtomic_StoreInt32(ptr, value)
}

func InternalRuntimeAtomic_Storeint64(ptr *int64, value int64) {
	SyncAtomic_StoreInt64(ptr, value)
}

func InternalRuntimeAtomic_StorepNoWB(ptr unsafe.Pointer, value unsafe.Pointer) {
	// The host collector still needs a write barrier even though the simulated
	// runtime does not model collection.
	SyncAtomic_StorePointer((*unsafe.Pointer)(ptr), value)
}

func InternalRuntimeAtomic_And32(ptr *uint32, value uint32) uint32 {
	return SyncAtomic_AndUint32(ptr, value)
}

func InternalRuntimeAtomic_Or32(ptr *uint32, value uint32) uint32 {
	return SyncAtomic_OrUint32(ptr, value)
}

func InternalRuntimeAtomic_And64(ptr *uint64, value uint64) uint64 {
	return SyncAtomic_AndUint64(ptr, value)
}

func InternalRuntimeAtomic_Or64(ptr *uint64, value uint64) uint64 {
	return SyncAtomic_OrUint64(ptr, value)
}

func InternalRuntimeAtomic_Anduintptr(ptr *uintptr, value uintptr) uintptr {
	return SyncAtomic_AndUintptr(ptr, value)
}

func InternalRuntimeAtomic_Oruintptr(ptr *uintptr, value uintptr) uintptr {
	return SyncAtomic_OrUintptr(ptr, value)
}

func InternalRuntimeAtomic_And(ptr *uint32, value uint32) {
	InternalRuntimeAtomic_And32(ptr, value)
}

func InternalRuntimeAtomic_Or(ptr *uint32, value uint32) {
	InternalRuntimeAtomic_Or32(ptr, value)
}

func InternalRuntimeAtomic_Load8(ptr *uint8) uint8 {
	runtimeAtomicByte()
	return *ptr
}

func InternalRuntimeAtomic_Store8(ptr *uint8, value uint8) {
	runtimeAtomicByte()
	*ptr = value
}

func InternalRuntimeAtomic_Xchg8(ptr *uint8, value uint8) uint8 {
	runtimeAtomicByte()
	old := *ptr
	*ptr = value
	return old
}

func InternalRuntimeAtomic_And8(ptr *uint8, value uint8) {
	runtimeAtomicByte()
	*ptr &= value
}

func InternalRuntimeAtomic_Or8(ptr *uint8, value uint8) {
	runtimeAtomicByte()
	*ptr |= value
}

// Pointer stores still need the host collector's write barrier, and must share
// the scheduling boundary used by the other simulated atomic operations.
func InternalRuntimeAtomic_storePointer(ptr *unsafe.Pointer, value unsafe.Pointer) {
	SyncAtomic_StorePointer(ptr, value)
}

// A simulated process has no host ELF auxiliary vector. Returning no entries
// makes unix.Auxv report its documented ENOENT result without exposing host
// addresses, CPU capabilities, or process initialization data.
func GolangOrgXSysUnix_runtime_getAuxv() []uintptr {
	return nil
}

func InternalRuntimeAtomic_casPointer(ptr *unsafe.Pointer, old, value unsafe.Pointer) bool {
	return SyncAtomic_CompareAndSwapPointer(ptr, old, value)
}

// The simulation owns panic reporting and cannot redirect the host runtime's
// crash descriptor. Disabling an output that has never been enabled is safe.
func RuntimeDebug_runtime_setCrashFD(fd uintptr) uintptr {
	if fd == ^uintptr(0) {
		return ^uintptr(0)
	}
	panic("gosim: runtime/debug.SetCrashOutput is unsupported; use simulation logs for panic reports")
}

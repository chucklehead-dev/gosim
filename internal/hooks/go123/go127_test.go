//go:build !sim

package go123

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func bitwiseMatches[T ~int32 | ~int64 | ~uint32 | ~uint64 | ~uintptr](t *testing.T, simulated, native func(*T, T) T) {
	t.Helper()
	for _, start := range []T{0, 1, 0x35, ^T(0)} {
		for _, mask := range []T{0, 0x16, ^T(0)} {
			gotValue, wantValue := start, start
			gotOld, wantOld := simulated(&gotValue, mask), native(&wantValue, mask)
			if gotOld != wantOld || gotValue != wantValue {
				t.Fatalf("start=%v mask=%v: old/value=%v/%v; native=%v/%v", start, mask, gotOld, gotValue, wantOld, wantValue)
			}
		}
	}
}

// Both public and new internal atomics must return the pre-operation value.
// Checking only the stored result misses an observable Go contract violation.
func TestBitwiseAtomicsMatchGo(t *testing.T) {
	t.Run("AndInt32", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_AndInt32, atomic.AndInt32) })
	t.Run("AndInt64", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_AndInt64, atomic.AndInt64) })
	t.Run("AndUint32", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_AndUint32, atomic.AndUint32) })
	t.Run("AndUint64", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_AndUint64, atomic.AndUint64) })
	t.Run("AndUintptr", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_AndUintptr, atomic.AndUintptr) })
	t.Run("OrInt32", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_OrInt32, atomic.OrInt32) })
	t.Run("OrInt64", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_OrInt64, atomic.OrInt64) })
	t.Run("OrUint32", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_OrUint32, atomic.OrUint32) })
	t.Run("OrUint64", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_OrUint64, atomic.OrUint64) })
	t.Run("OrUintptr", func(t *testing.T) { bitwiseMatches(t, SyncAtomic_OrUintptr, atomic.OrUintptr) })
	t.Run("internalAnd32", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_And32, atomic.AndUint32) })
	t.Run("internalAnd64", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_And64, atomic.AndUint64) })
	t.Run("internalAnduintptr", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_Anduintptr, atomic.AndUintptr) })
	t.Run("internalOr32", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_Or32, atomic.OrUint32) })
	t.Run("internalOr64", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_Or64, atomic.OrUint64) })
	t.Run("internalOruintptr", func(t *testing.T) { bitwiseMatches(t, InternalRuntimeAtomic_Oruintptr, atomic.OrUintptr) })
}

func TestRuntimeAtomicsPreserveValues(t *testing.T) {
	var value uint64 = 8
	if got := InternalRuntimeAtomic_Xadd64(&value, -3); got != 5 || value != 5 {
		t.Fatalf("add: result=%d value=%d", got, value)
	}
	if got := InternalRuntimeAtomic_Xchg64(&value, 12); got != 5 || value != 12 {
		t.Fatalf("swap: old=%d value=%d", got, value)
	}
	if InternalRuntimeAtomic_Cas64(&value, 7, 9) || !InternalRuntimeAtomic_Cas64(&value, 12, 9) || value != 9 {
		t.Fatalf("compare-and-swap value=%d", value)
	}
	InternalRuntimeAtomic_StoreRel64(&value, 32)
	if got := InternalRuntimeAtomic_LoadAcq64(&value); got != 32 {
		t.Fatalf("release/acquire value=%d", got)
	}
	var target, replacement int
	p := unsafe.Pointer(&target)
	if !InternalRuntimeAtomic_Casp1(&p, unsafe.Pointer(&target), unsafe.Pointer(&replacement)) || InternalRuntimeAtomic_Loadp(unsafe.Pointer(&p)) != unsafe.Pointer(&replacement) {
		t.Fatal("pointer compare-and-swap/load")
	}
	InternalRuntimeAtomic_StorepNoWB(unsafe.Pointer(&p), unsafe.Pointer(&target))
	if p != unsafe.Pointer(&target) {
		t.Fatal("pointer store")
	}
	InternalRuntimeAtomic_storePointer(&p, unsafe.Pointer(&replacement))
	if p != unsafe.Pointer(&replacement) {
		t.Fatal("write-barrier pointer store")
	}
}

func TestFIPSTimingRefusesHostEntropy(t *testing.T) {
	defer func() {
		p := recover()
		message, ok := p.(string)
		if !ok || !strings.Contains(message, "FIPS CPU-jitter entropy timing is unsupported") {
			t.Fatalf("expected explicit unsupported FIPS timing, got %v", p)
		}
	}()
	CryptoInternalFips140depsTime_monoTime()
}

func TestAuxvDoesNotExposeHostProcess(t *testing.T) {
	if got := GolangOrgXSysUnix_runtime_getAuxv(); len(got) != 0 {
		t.Fatalf("simulated process has no ELF auxv, got %v", got)
	}
}

func TestRuntimePointerCASMatchesGo(t *testing.T) {
	var first, second, third int
	for _, replacement := range []*int{&second, nil, &third} {
		for _, match := range []bool{false, true} {
			got, want := unsafe.Pointer(&first), unsafe.Pointer(&first)
			old := unsafe.Pointer(&first)
			if !match {
				old = unsafe.Pointer(&second)
			}
			gotSwapped := InternalRuntimeAtomic_casPointer(&got, old, unsafe.Pointer(replacement))
			wantSwapped := atomic.CompareAndSwapPointer(&want, old, unsafe.Pointer(replacement))
			if gotSwapped != wantSwapped || got != want {
				t.Fatalf("CAS match=%v: swapped/value=%v/%p; native=%v/%p", match, gotSwapped, got, wantSwapped, want)
			}
		}
	}
}

func TestNativeDueTimerSelectReadiness(t *testing.T) {
	for _, delay := range []time.Duration{0, -time.Second} {
		for range 100 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			default:
				t.Fatalf("native due timer (%v) was not ready", delay)
			}
			if timer.Stop() {
				t.Fatal("native received timer remained active")
			}
		}
	}
	timer := time.NewTimer(time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	select {
	case <-timer.C:
	default:
		t.Fatal("native elapsed timer was not ready")
	}
}

func TestCrashOutputDoesNotModifyHostRuntime(t *testing.T) {
	if previous := RuntimeDebug_runtime_setCrashFD(^uintptr(0)); previous != ^uintptr(0) {
		t.Fatalf("unexpected previous crash descriptor %d", previous)
	}
	defer func() {
		message, ok := recover().(string)
		if !ok || !strings.Contains(message, "runtime/debug.SetCrashOutput is unsupported") {
			t.Fatalf("expected actionable unsupported crash output, got %q", message)
		}
	}()
	RuntimeDebug_runtime_setCrashFD(123)
}

func TestRuntimeTraceRejectsHostRecording(t *testing.T) {
	for name, invoke := range map[string]func(){
		"advance":   func() { RuntimeTrace_runtime_traceAdvance(false) },
		"clock":     func() { RuntimeTrace_runtime_traceClockNow() },
		"frequency": func() { RuntimeTrace_runtime_traceClockUnitsPerSecond() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				message, ok := recover().(string)
				if !ok || !strings.Contains(message, "runtime/trace recording is unsupported") {
					t.Fatalf("expected actionable unsupported runtime trace, got %q", message)
				}
			}()
			invoke()
		})
	}
}

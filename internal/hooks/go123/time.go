package go123

import (
	"time"
	"unsafe"

	"github.com/chucklehead-dev/gosim/gosimruntime"
	"github.com/chucklehead-dev/gosim/internal/race"
)

func Time_now() (sec int64, nsec int32, mono int64) {
	now := gosimruntime.Nanotime()

	// TODO: somehow avoid this division?
	sec = now / 1e9
	mono = now
	nsec = int32(mono - sec*1e9)

	return
}

func Time_runtimeNano() int64 {
	return gosimruntime.Nanotime()
}

func Time_Sleep(duration time.Duration) {
	gosimruntime.Sleep(int64(duration))
}

//go:norace
func fireRuntimeTimer(t *gosimruntime.Timer) {
	rt := t.Arg.(*timeTimer)
	machine := t.Machine

	when := rt.when
	if rt.c != nil {
		// Channel readiness must be established before the scheduler resumes
		// another goroutine. Capture the time base in the creating machine,
		// since time.Now accesses translated machine globals.
		value := rt.base.Add(time.Duration(when - rt.base.UnixNano()))
		gosimruntime.TimerChan[time.Time](rt.c).SendTimer(value, rt.seq)
	} else {
		f, arg, seq := rt.f, rt.arg, rt.seq
		gosimruntime.GoFromTimerHandler(func() {
			if race.Enabled {
				race.Acquire(unsafe.Pointer(rt))
			}
			f(arg, seq, gosimruntime.Nanotime()-when)
		}, machine)
	}

	if rt.period != 0 {
		// FIXME: this might be borked; base off now?
		rt.when += rt.period
		t.Reset(rt.when)
	}
}

type timeTimer struct {
	c    unsafe.Pointer
	init bool
	timer
	base time.Time
}

type timer struct {
	pp     *gosimruntime.Timer // repurposed pp
	when   int64
	period int64
	f      func(arg any, seq uintptr, delta int64) // NOTE: must not be closure
	arg    any                                     // passed to f
	seq    uintptr                                 // passed to f
}

//go:norace
func Time_newTimer(when, period int64, f func(any, uintptr, int64), arg any, cp unsafe.Pointer) *timeTimer {
	t := &timeTimer{
		init: true, // huh
	}

	if race.Enabled {
		race.Release(unsafe.Pointer(t))
	}

	t.c = cp
	t.base = time.Now()
	if cp != nil {
		gosimruntime.TimerChan[time.Time](cp).ConfigureTimer(t.seq, period != 0)
	}

	t.when = when
	t.period = period
	t.f = f
	t.arg = arg
	t.seq = 0

	t.pp = gosimruntime.NewTimer(fireRuntimeTimer, t, gosimruntime.CurrentMachine(), t.when)

	return t
}

//go:norace
func Time_stopTimer(t *timeTimer) bool {
	stopped := t.pp.Stop()
	if t.c != nil {
		t.seq++
		return gosimruntime.TimerChan[time.Time](t.c).StopTimer(t.seq)
	}
	return stopped
}

//go:norace
func Time_resetTimer(t *timeTimer, when, period int64) bool {
	if race.Enabled {
		race.Release(unsafe.Pointer(t))
	}

	t.when = when
	t.period = period
	t.base = time.Now()
	stopped := t.pp.Reset(when)
	if t.c != nil {
		t.seq++
		return gosimruntime.TimerChan[time.Time](t.c).ConfigureTimer(t.seq, period != 0)
	}
	return stopped
}

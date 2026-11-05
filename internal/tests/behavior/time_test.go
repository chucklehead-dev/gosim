//go:build sim

package behavior_test

import (
	"log"
	"log/slog"
	"testing"
	"time"
)

// TODO: test these against real behavior with some slack built in for time
// TODO: test monotonic time with different machines/universes

func TestTimeBasic(t *testing.T) {
	a := time.Now()

	time.Sleep(10 * time.Second)

	b := time.Now()
	if b.Sub(a) < 10*time.Second {
		t.Error(a, b)
	}

	time.Sleep(10 * time.Second)

	c := time.Now()
	if c.Sub(b) < 10*time.Second {
		t.Error(a, b)
	}

	log.Println(a, b, c)
}

func TestTimerPast(t *testing.T) {
	a := time.Now()
	timer := time.NewTimer(-10 * time.Second)
	<-timer.C
	b := time.Now()
	if !a.Equal(b) {
		t.Error(a, b)
	}
}

func TestTimerAfter(t *testing.T) {
	a := time.Now()

	<-time.After(10 * time.Second)

	b := time.Now()
	if b.Sub(a) != 10*time.Second {
		t.Error(a, b)
	}

	<-time.After(10 * time.Second)

	c := time.Now()
	if c.Sub(b) != 10*time.Second {
		t.Error(a, b)
	}
}

func TestTimerValue(t *testing.T) {
	a := time.Now()
	ret := <-time.After(10 * time.Second)
	b := time.Now()
	if b.Sub(a) != 10*time.Second {
		t.Error(a, b)
	}
	if !ret.Equal(b) {
		t.Error(ret, b)
	}
}

func TestTimerValueSleep(t *testing.T) {
	a := time.Now()
	ch := time.After(10 * time.Second)
	time.Sleep(20 * time.Second)
	ret := <-ch
	b := time.Now()
	if b.Sub(a) != 20*time.Second {
		t.Error(a, b)
	}
	if ret.Sub(a) != 10*time.Second {
		t.Error(ret, a)
	}
}

func TestTimerResetOk(t *testing.T) {
	a := time.Now()

	timer := time.NewTimer(10 * time.Second)
	if !timer.Reset(20 * time.Second) {
		t.Error("bad reset")
	}
	<-timer.C

	b := time.Now()
	if b.Sub(a) != 20*time.Second {
		t.Error(a, b)
	}
}

func TestTimerStopOk(t *testing.T) {
	a := time.Now()

	timer := time.NewTimer(10 * time.Second)

	if timer.Stop() {
		timer.Reset(20 * time.Second)
	} else {
		t.Error("bad")
	}
	<-timer.C

	b := time.Now()
	if b.Sub(a) != 20*time.Second {
		t.Error(a, b)
	}
}

// An expired but unread channel timer is still stoppable. Stop removes its
// pending value, and Reset cannot expose that value as a stale receive.
func TestTimerStopExpiredUnread(t *testing.T) {
	a := time.Now()
	timer := time.NewTimer(10 * time.Second)
	if cap(timer.C) != 0 || len(timer.C) != 0 {
		t.Fatal("timer channel must appear synchronous")
	}
	time.Sleep(20 * time.Second)
	if len(timer.C) != 0 || !timer.Stop() {
		t.Fatal("expired unread timer was not stopped")
	}
	select {
	case <-timer.C:
		t.Fatal("Stop left a stale timer value")
	default:
	}
	if timer.Stop() || timer.Reset(20*time.Second) {
		t.Fatal("stopped timer was still active")
	}
	value := <-timer.C
	if value.Sub(a) != 40*time.Second || time.Since(a) != 40*time.Second {
		t.Fatal("Reset delivered an old value", value, a)
	}
	if timer.Stop() {
		t.Fatal("received timer was still active")
	}
}

func TestTimerResetExpiredUnread(t *testing.T) {
	a := time.Now()
	timer := time.NewTimer(10 * time.Second)
	time.Sleep(20 * time.Second)
	if !timer.Reset(5 * time.Second) {
		t.Fatal("unread timer Reset must report active")
	}
	select {
	case <-timer.C:
		t.Fatal("Reset left a stale timer value")
	default:
	}
	if value := <-timer.C; value.Sub(a) != 25*time.Second {
		t.Fatal("Reset delivered an old value", value, a)
	}
}

func TestTimerStopImmediate(t *testing.T) {
	// The timer is ready at the current instant, but its callback may only be
	// queued. Stop must invalidate that callback as well as any hidden value.
	for range 20 {
		timer := time.NewTimer(0)
		if !timer.Stop() {
			t.Fatal("unreceived immediate timer was not stopped")
		}
		time.Sleep(time.Nanosecond)
		select {
		case <-timer.C:
			t.Fatal("queued callback sent after Stop")
		default:
		}
	}
}

func TestTimerResetImmediate(t *testing.T) {
	for range 20 {
		start := time.Now()
		timer := time.NewTimer(0)
		if !timer.Reset(time.Second) {
			t.Fatal("unreceived immediate timer was not active")
		}
		if value := <-timer.C; value.Sub(start) != time.Second {
			t.Fatal("queued callback sent after Reset", value, start)
		}
	}
}

func TestTimerAfterFunc(t *testing.T) {
	a := time.Now()

	done := make(chan struct{}, 0)
	time.AfterFunc(10*time.Second, func() {
		done <- struct{}{}
	})

	<-done
	b := time.Now()
	if b.Sub(a) != 10*time.Second {
		t.Error(a, b)
	}
}

func TestTimerAfterFuncReset(t *testing.T) {
	a := time.Now()

	done := make(chan struct{}, 0)
	timer := time.AfterFunc(10*time.Second, func() {
		done <- struct{}{}
	})

	<-done
	b := time.Now()
	if b.Sub(a) != 10*time.Second {
		t.Error(a, b)
	}

	timer.Reset(10 * time.Second)
	timer.Reset(5 * time.Second)

	<-done
	c := time.Now()
	if c.Sub(b) != 5*time.Second {
		t.Error(b, c)
	}
}

func TestTimeSameStart(t *testing.T) {
	time.Sleep(10 * time.Second)

	// TODO: need a test for determinism here?
	slog.Info("after", "now", time.Now().String())
}

func TestTimeTickerStop(t *testing.T) {
	start := time.Now()

	ticker := time.NewTicker(5 * time.Second)

	v := <-ticker.C
	if now := time.Now(); now.Sub(start) != 5*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 5*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 10*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 10*time.Second {
		t.Error(start, v)
	}

	ticker.Stop()
	time.Sleep(10 * time.Second)
	select {
	case v := <-ticker.C:
		t.Error(v, start)
	default:
	}
}

func TestTimeTickerReset(t *testing.T) {
	start := time.Now()

	ticker := time.NewTicker(5 * time.Second)

	v := <-ticker.C
	if now := time.Now(); now.Sub(start) != 5*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 5*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 10*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 10*time.Second {
		t.Error(start, v)
	}

	time.Sleep(2 * time.Second)
	ticker.Reset(6 * time.Second)

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 18*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 18*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 24*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 24*time.Second {
		t.Error(start, v)
	}

	time.Sleep(2 * time.Second)
	ticker.Reset(2 * time.Second)

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 28*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 28*time.Second {
		t.Error(start, v)
	}

	time.Sleep(7 * time.Second)
	ticker.Reset(5 * time.Second)

	// Reset removes the old pending tick at 30: the next tick is at 40.
	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 40*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 40*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 45*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 45*time.Second {
		t.Error(start, v)
	}
}

func TestTimeTickerStopReset(t *testing.T) {
	start := time.Now()

	ticker := time.NewTicker(5 * time.Second)

	v := <-ticker.C
	if now := time.Now(); now.Sub(start) != 5*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 5*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 10*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 10*time.Second {
		t.Error(start, v)
	}

	ticker.Stop()
	time.Sleep(10 * time.Second)
	select {
	case v := <-ticker.C:
		t.Error(v, start)
	default:
	}

	ticker.Reset(3 * time.Second)
	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 23*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 23*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 26*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 26*time.Second {
		t.Error(start, v)
	}
}

func TestTimeTickerNonBlocking(t *testing.T) {
	start := time.Now()

	ticker := time.NewTicker(5 * time.Second)

	v := <-ticker.C
	if now := time.Now(); now.Sub(start) != 5*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 5*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 10*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 10*time.Second {
		t.Error(start, v)
	}

	time.Sleep(8 * time.Second)
	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 18*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 15*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 20*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 20*time.Second {
		t.Error(start, v)
	}

	time.Sleep(31 * time.Second)
	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 51*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 25*time.Second {
		t.Error(start, v)
	}

	v = <-ticker.C
	if now := time.Now(); now.Sub(start) != 55*time.Second {
		t.Error(start, now)
	}
	if v.Sub(start) != 55*time.Second {
		t.Error(start, v)
	}
}

// Channel operations must observe due timers immediately, without waiting for
// a separately scheduled delivery goroutine. This matches Go's maybeRunChan.
func TestDueTimerSelectReadiness(t *testing.T) {
	for _, delay := range []time.Duration{0, -time.Second} {
		for range 50 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			default:
				t.Fatalf("due timer (%v) was not ready", delay)
			}
			if timer.Stop() {
				t.Fatal("received timer remained active")
			}
		}
	}
	start := time.Now()
	timer := time.NewTimer(time.Second)
	time.Sleep(time.Second)
	select {
	case value := <-timer.C:
		if value.Sub(start) != time.Second {
			t.Fatal("incorrect deadline", value, start)
		}
	default:
		t.Fatal("timer was not ready at its deadline")
	}
	// Reset and Stop must discard readiness established at a prior deadline.
	timer.Reset(0)
	timer.Reset(time.Second)
	select {
	case <-timer.C:
		t.Fatal("Reset exposed old readiness")
	default:
	}
	timer.Reset(0)
	if !timer.Stop() {
		t.Fatal("unreceived timer was inactive")
	}
	select {
	case <-timer.C:
		t.Fatal("Stop exposed old readiness")
	default:
	}
}

//go:build !sim

package behavior_test

import (
	"testing"
	"testing/synctest"
	"time"
)

// Native Go is the contract oracle for the equivalent virtual-time timer tests.
// synctest runs only this host test, never inside a gosim simulation.
func TestNativeTimerChannelContracts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		timer := time.NewTimer(time.Second)
		if len(timer.C) != 0 || cap(timer.C) != 0 {
			t.Fatal("timer channel must be synchronous")
		}
		time.Sleep(2 * time.Second)
		if !timer.Stop() {
			t.Fatal("expired unread timer must remain stoppable")
		}
		select {
		case <-timer.C:
			t.Fatal("stale value after Stop")
		default:
		}
		if timer.Stop() || timer.Reset(3*time.Second) {
			t.Fatal("stopped timer reported active")
		}
		if got := <-timer.C; got.Sub(start) != 5*time.Second {
			t.Fatal("unexpected reset value", got)
		}
		if timer.Stop() {
			t.Fatal("received timer reported active")
		}

		timer.Reset(time.Second)
		time.Sleep(2 * time.Second)
		if !timer.Reset(3 * time.Second) {
			t.Fatal("expired unread timer Reset must report active")
		}
		select {
		case <-timer.C:
			t.Fatal("stale value after Reset")
		default:
		}
		if got := <-timer.C; got.Sub(start) != 10*time.Second {
			t.Fatal("unexpected reset value", got)
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		time.Sleep(2 * time.Second)
		ticker.Reset(3 * time.Second)
		select {
		case <-ticker.C:
			t.Fatal("stale tick after Reset")
		default:
		}
		if got := <-ticker.C; got.Sub(start) != 15*time.Second {
			t.Fatal("unexpected ticker value", got)
		}
	})
}

package gosimruntime

import (
	"os"
	"testing"
)

func TestTimerDeliveryOnScheduler(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		delivered := false
		run(func() {
			c := NewChan[int](1)
			c.ConfigureTimer(1, false)
			NewTimer(func(*Timer) { c.SendTimer(42, 1) }, nil, CurrentMachine(), Nanotime()+1)
			// Blocked receivers must be resumed directly by scheduler delivery.
			delivered = c.Recv() == 42
		}, seed, false, false, "INFO", makeConsoleLogger(os.Stderr), nil)
		if !delivered {
			t.Fatalf("seed %d: timer did not deliver to blocked receiver", seed)
		}
	}
}

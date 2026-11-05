//go:build !sim

package behavior_test

import (
	"testing"

	"github.com/chucklehead-dev/gosim/metatesting"
)

// Reusing a runner would miss accidental dependencies on host-process hash keys
// and addresses. Start a new executable for each sample, then reuse that runner
// once to verify per-run identities reset too.
func TestMetaMapHashFreshProcess(t *testing.T) {
	for _, test := range []string{"TestRandMapHashCoverage", "TestRandMapHashComparable"} {
		t.Run(test, func(t *testing.T) {
			var first string
			for process := range 3 {
				runner, err := metatesting.Start(t, "github.com/chucklehead-dev/gosim/internal/tests/behavior")
				if err != nil {
					t.Fatal(err)
				}
				for range 2 {
					result, err := runner.Run(t, &metatesting.RunConfig{Test: test, Seed: 123})
					if err != nil {
						runner.Close()
						t.Fatal(err)
					}
					if result.Failed {
						runner.Close()
						t.Fatalf("simulation failed: %s", result.Err)
					}
					value := metatesting.MustFindLogValue(metatesting.ParseLog(result.LogOutput), "rand", "val").(string)
					if process == 0 && first == "" {
						first = value
					}
					if value != first {
						runner.Close()
						t.Fatalf("hash changed across runs/processes: %s != %s", value, first)
					}
				}
				runner.Close()
			}
		})
	}
}

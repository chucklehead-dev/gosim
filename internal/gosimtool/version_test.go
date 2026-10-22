package gosimtool

import (
	"strings"
	"testing"
)

func TestCheckGoVersion(t *testing.T) {
	for _, v := range []string{"go1.26", "go1.26.1", "go1.26rc2", "devel go1.27-abc"} {
		if err := CheckGoVersion(v); err != nil {
			t.Errorf("%s: %v", v, err)
		}
	}
	for _, v := range []string{"go1.25.3", "go1.27.0", "go1.260"} {
		err := CheckGoVersion(v)
		if err == nil || !strings.Contains(err.Error(), "GOTOOLCHAIN=go1.26.1") {
			t.Errorf("%s: %v, want the fix", v, err)
		}
	}
}

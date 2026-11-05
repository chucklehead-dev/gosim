package translate

import (
	"testing"

	"github.com/chucklehead-dev/gosim/internal/gosimtool"
)

func TestTranslateCacheIncludesEffectiveToolchain(t *testing.T) {
	base := gosimtool.BuildConfig{GOOS: "linux", GOARCH: "amd64", Toolchain: gosimtool.Toolchain{Version: "go1.27.3", GOROOT: "/toolchains/current"}}
	original := computeTranslateToolHash(base)
	if again := computeTranslateToolHash(base); again != original {
		t.Fatal("unchanged effective toolchain changed the cache key")
	}
	for _, cfg := range []gosimtool.BuildConfig{
		{GOOS: "linux", GOARCH: "amd64", Toolchain: gosimtool.Toolchain{Version: "go1.27.4", GOROOT: base.Toolchain.GOROOT}},
		{GOOS: "linux", GOARCH: "amd64", Toolchain: gosimtool.Toolchain{Version: base.Toolchain.Version, GOROOT: "/toolchains/other"}},
	} {
		if computeTranslateToolHash(cfg) == original {
			t.Fatalf("different effective toolchain reused translated cache: %+v", cfg.Toolchain)
		}
	}
}

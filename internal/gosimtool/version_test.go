package gosimtool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestCheckGoVersion(t *testing.T) {
	for _, v := range []string{"go1.27", "go1.27.0", "go1.27.3"} {
		if err := CheckGoVersion(v); err != nil {
			t.Errorf("%s: %v", v, err)
		}
	}
	for _, v := range []string{"go1.26.10", "go1.28.0", "go1.270", "go1.27rc2", "go1.27beta1", "devel go1.27-abc", ""} {
		err := CheckGoVersion(v)
		if err == nil || !strings.Contains(err.Error(), "GOTOOLCHAIN=go1.27.3") {
			t.Errorf("%s: %v, want the supported toolchain fix", v, err)
		}
	}
}

func TestResolveToolchainUsesSelectedGoAndRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Go launcher is a Unix shell script")
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("go1.27.3\ntime 2026-11-01T00:00:00Z\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	launcher := "#!/bin/sh\n[ \"$*\" = 'env -json GOVERSION GOROOT' ] || exit 77\nprintf '%s' \"$GOSIM_TEST_GO_ENV\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(launcher), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, tc := range []struct {
		name, version, goroot, want string
	}{
		{"supported", "go1.27.3", root, ""},
		// An installed gosim built by the supported compiler must still reject
		// another go command's selected standard-library family.
		{"different launcher family", "go1.26.10", root, "not go1.26.10"},
		{"mismatched GOROOT", "go1.27.4", root, "contains go1.27.3"},
		{"relative root", "go1.27.3", "relative", "not absolute"},
		{"unknown development compiler", "devel go1.27-abc", root, "not devel"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(map[string]string{"GOVERSION": tc.version, "GOROOT": tc.goroot})
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("GOSIM_TEST_GO_ENV", string(data))
			toolchain, err := ResolveToolchain()
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("selected go = %+v, %v; want %q", toolchain, err, tc.want)
				}
				return
			}
			if err != nil || toolchain.Version != tc.version || toolchain.GOROOT != root {
				t.Fatalf("selected go = %+v, %v", toolchain, err)
			}
			if !slices.Equal(toolchain.Env(), []string{"GOTOOLCHAIN=go1.27.3", "GOROOT=" + root}) {
				t.Fatalf("unpinned subprocess selection: %v", toolchain.Env())
			}
		})
	}
}

func TestResolveToolchainRejectsInvalidGoEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Go launcher is a Unix shell script")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\nprintf 'not json'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if _, err := ResolveToolchain(); err == nil || !strings.Contains(err.Error(), "invalid go env output") {
		t.Fatalf("invalid environment: %v", err)
	}
}

func TestResolveToolchainPreservesGoSelectionFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Go launcher is a Unix shell script")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte("#!/bin/sh\nprintf 'go.mod requires go >= 1.27.3' >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if _, err := ResolveToolchain(); err == nil || !strings.Contains(err.Error(), "requires go >= 1.27.3") {
		t.Fatalf("lost actionable Go selection error: %v", err)
	}
}

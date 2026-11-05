package gosimtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/version"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/mod/modfile"

	"github.com/chucklehead-dev/gosim/internal/race"
)

const (
	Module          = "github.com/chucklehead-dev/gosim"
	OutputDirectory = ".gosim"
)

// SupportedGo is the Go release whose standard library the translator hooks
// target. Other release families are rejected before translation.
const SupportedGo = "go1.27"

// RecommendedGo is the maintained toolchain used to validate this target.
const RecommendedGo = "go1.27.3"

// CheckGoVersion accepts stable releases of the supported family. Unknown
// development and prerelease runtimes must not silently claim hook support.
func CheckGoVersion(v string) error {
	if version.IsValid(v) && version.Lang(v) == SupportedGo && !strings.Contains(v, "rc") && !strings.Contains(v, "beta") {
		return nil
	}
	return fmt.Errorf("this gosim translates %s's standard library, not %s's: select GOTOOLCHAIN=%s or use a gosim release for %s", SupportedGo, v, RecommendedGo, v)
}

// Toolchain identifies the effective compiler and standard library selected by
// the go command, including GOTOOLCHAIN selection in the caller's module.
type Toolchain struct {
	Version string
	GOROOT  string
}

// ResolveToolchain validates the go command's actual selection, rather than the
// runtime that happened to build an installed gosim executable.
func ResolveToolchain() (Toolchain, error) {
	cmd := exec.Command("go", "env", "-json", "GOVERSION", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		if failure, ok := errors.AsType[*exec.ExitError](err); ok && len(failure.Stderr) > 0 {
			return Toolchain{}, fmt.Errorf("resolve Go toolchain: %w: %s", err, strings.TrimSpace(string(failure.Stderr)))
		}
		return Toolchain{}, fmt.Errorf("resolve Go toolchain: %w", err)
	}
	var selection struct{ GOVERSION, GOROOT string }
	if err := json.Unmarshal(output, &selection); err != nil {
		return Toolchain{}, fmt.Errorf("resolve Go toolchain: invalid go env output: %w", err)
	}
	if err := CheckGoVersion(selection.GOVERSION); err != nil {
		return Toolchain{}, err
	}
	if !filepath.IsAbs(selection.GOROOT) {
		return Toolchain{}, fmt.Errorf("resolve Go toolchain: GOROOT %q is not absolute", selection.GOROOT)
	}
	root, err := filepath.EvalSymlinks(selection.GOROOT)
	if err != nil {
		return Toolchain{}, fmt.Errorf("resolve GOROOT: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return Toolchain{}, fmt.Errorf("read GOROOT version: %w", err)
	}
	rootVersion, _, _ := strings.Cut(string(data), "\n")
	if rootVersion != selection.GOVERSION {
		return Toolchain{}, fmt.Errorf("go env selects %s but GOROOT %s contains %s: clear a mismatched GOROOT", selection.GOVERSION, root, rootVersion)
	}
	return Toolchain{Version: selection.GOVERSION, GOROOT: root}, nil
}

// Env pins subprocess package loading/builds to this validated selection.
func (t Toolchain) Env() []string {
	return []string{"GOTOOLCHAIN=" + t.Version, "GOROOT=" + t.GOROOT}
}

type BuildConfig struct {
	Toolchain Toolchain
	GOOS      string
	GOARCH    string
	Race      bool
}

func (b BuildConfig) AsDirname() string {
	name := fmt.Sprintf("%s_%s", b.GOOS, b.GOARCH)
	if b.Race {
		name += "_race"
	}
	return name
}

type TranslateOutput struct {
	RootOutputDir string
	Packages      []string                        // TODO: combine all this in a struct? also include input name?
	Deps          map[string]map[string]time.Time // pkg -> path -> time
}

func FindGoMod() (string, *modfile.File, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	for i := 0; i < 20; i++ {
		p := path.Join(baseDir, "go.mod")
		bytes, err := os.ReadFile(p)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return "", nil, err
			}

			baseDir = path.Dir(baseDir)
			if baseDir != "." {
				continue
			}
			return "", nil, errors.New("could not find go.mod file, hit root")
		}

		file, err := modfile.Parse(p, bytes, nil)
		if err != nil {
			return "", nil, err
		}

		return p, file, nil
	}

	return "", nil, errors.New("could not find go.mod file, tried too many dirs")
}

func FindGoModDir() (string, error) {
	modFile, _, err := FindGoMod()
	if err != nil {
		return "", err
	}

	modDir := filepath.Dir(modFile)
	return modDir, nil
}

// MakeGoModForTest makes a go.mod and go.sum for tests that use the `gosim` CLI
// outside of the gosim module. It takes the go.mod from the gosim module strips
// it, and adds a `replace` directive pointing to the gosim module.
//
// Only dependencies that are listed in the original go.mod work.
func MakeGoModForTest(goSimModDir, testModDir string, requires []string) error {
	origModPath := filepath.Join(goSimModDir, "go.mod")
	origModBytes, err := os.ReadFile(origModPath)
	if err != nil {
		return err
	}

	origFile, err := modfile.Parse(origModPath, origModBytes, nil)
	if err != nil {
		return err
	}

	origSumBytes, err := os.ReadFile(filepath.Join(goSimModDir, "go.sum"))
	if err != nil {
		return err
	}

	origRequires := make(map[string]*modfile.Require)
	for _, mod := range origFile.Require {
		origRequires[mod.Mod.Path] = mod
	}

	newFile := &modfile.File{}
	newFile.AddModuleStmt("test")
	newFile.AddRequire(origFile.Module.Mod.Path, "v1.0.0")
	newFile.AddGoStmt(origFile.Go.Version)
	for _, pkg := range requires {
		require, ok := origRequires[pkg]
		if !ok {
			return fmt.Errorf("could not find pkg %s", pkg)
		}
		newFile.AddRequire(require.Mod.Path, require.Mod.Version)
	}
	newFile.AddReplace(origFile.Module.Mod.Path, "", goSimModDir, "")
	newModBytes, err := newFile.Format()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(testModDir, "go.mod"), newModBytes, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(testModDir, "go.sum"), origSumBytes, 0o644); err != nil {
		return err
	}
	return nil
}

func isUncachedTestRun() bool {
	// TODO: fallback only allowed if -count=1 (which bans caching)
	flags := []string{"-test.count="}

	for _, arg := range os.Args {
		for _, flag := range flags {
			// if we have this flag and it looks non-empty, assume non-cached test run
			if strings.HasPrefix(arg, flag) && len(arg) > len(flag) {
				return true
			}
		}
	}

	return false
}

func ReplaceSpecialPackages(s string) string {
	changed := false
	parts := strings.Split(s, "/")
	for i, part := range parts {
		if part == "vendor" || part == "internal" {
			parts[i] = part + "_"
			changed = true
		}
	}
	if !changed {
		return s
	}
	return strings.Join(parts, "/")
}

func PreparedTestBinName(pkg string) string {
	return strings.ReplaceAll(pkg, "/", "-") + ".test"
}

func PreparedTestInfoName(pkg string) string {
	return strings.ReplaceAll(pkg, "/", "-") + ".info"
}

func GetNewPackageName(pkg string) string {
	return "translated/" + ReplaceSpecialPackages(pkg)
}

var (
	buildOnceMu sync.Mutex
	buildOnce   = make(map[string]*sync.Once)
)

func GetPathForPrecompiledTestBinary(t *testing.T, pkg string) string {
	didBuild := false

	if isUncachedTestRun() {
		// TODO: figure out if we should pass -race or not?
		buildOnceMu.Lock()
		once, ok := buildOnce[pkg]
		if !ok {
			once = new(sync.Once)
			buildOnce[pkg] = once
		}
		buildOnceMu.Unlock()

		didBuild = true
		once.Do(func() {
			var name string
			var flags []string
			if prebuilt, ok := os.LookupEnv("GOSIMTOOL"); ok {
				name = prebuilt
			} else {
				name = "go"
				flags = []string{"run", Module + "/cmd/gosim"}
			}
			flags = append(flags, "build-tests")
			if race.Enabled {
				flags = append(flags, "-race")
			}
			flags = append(flags, pkg)
			cmd := exec.Command(name, flags...)
			t.Logf("looks like an uncached test run... running `%s`", strings.Join(cmd.Args, " "))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("gosim build-tests failed: %s\n\n%s", err, string(out))
			}
		})
	}

	goModDir, err := FindGoModDir()
	if err != nil {
		t.Fatalf("Could not find containing go module: %s", err)
	}

	// TODO: test this race stuff somehow. make test output if race is enabled?
	cfg := BuildConfig{
		GOOS:   "linux",
		GOARCH: runtime.GOARCH,
		Race:   race.Enabled,
	}

	outDir := filepath.Join(goModDir, OutputDirectory, "metatest", cfg.AsDirname())
	path := filepath.Join(outDir, PreparedTestBinName(GetNewPackageName(pkg)))
	infoPath := filepath.Join(outDir, PreparedTestInfoName(GetNewPackageName(pkg)))

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Could not find pre-built gosim test binary for package %q. Pre-building is necessary to support the go test cache. "+
				"Pre-build the test binary using `go run %s/cmd/gosim build-tests %s` or "+
				"run `go test` with `-count=1` to disable the go test cache and build during the test.", pkg, Module, pkg)
		}
	}

	if !didBuild {
		t.Logf("Using pre-built gosim test binary %q", path)

		// TODO: only do this once per binary?
		// TODO: maybe this is less than optimal... now if files get touched but the binary doesn't change
		// we still re-run tests... worthwhile trade-off to catch misuse?

		// check if the precompiled test binary is up-to-date "enough"
		bytes, err := os.ReadFile(infoPath)
		if err != nil {
			t.Fatalf("failed to read info file: %s", err)
		}
		var deps map[string]time.Time
		if err := json.Unmarshal(bytes, &deps); err != nil {
			t.Fatalf("failed to parse info file: %s", err)
		}
		for dep, ts := range deps {
			info, err := os.Stat(dep)
			if err != nil {
				t.Fatalf("failed to stat dependency %q: %s", dep, err)
			}
			if !ts.Equal(info.ModTime()) {
				t.Fatalf("Pre-built gosim test binary is out of date. Update binaries with gosim build-tests or run an uncached run with -count=1. See the documentation for metatesting. Input %q changed: %s vs %s.", dep, ts, info.ModTime())
			}
		}
	}

	return path
}

# gosim, forked for gotogether

This repository is [gosim](https://github.com/jellevandenhooff/gosim), by
Jelle van den Hooff (MIT, see LICENSE), at ffd3a61 with the changes
[gotogether](https://github.com/chucklehead-dev/gotogether) needs, kept as ordinary commits on top of
upstream's history so that they can be rebased onto a newer upstream:

- Go 1.27 port: hooks for the runtime entry points added or moved since
  Go 1.23 (`internal/hooks/go123/go126.go` and `go127.go`), the translator reporting every
  missing hook in one run, `purego` builds for the FIPS 140 crypto
  packages, reflect's newer API, `O_DIRECTORY`.
- A `Chooser` (`gosimruntime/choose.go`): every bounded choice the runtime
  makes (schedule, select, map order) and `gosimruntime.Choose` go through
  it; unbounded randomness comes from a generator seeded by one choice.
  The metatest protocol lets the runner be the chooser and receive the
  lines the program `Emit`s (`metatesting.MetaT.RunWith`), and runs
  binaries in parallel (`Start`, `Close`, `Alive`).

- Overlays (`internal/translate/weave.go`): `GOSIM_OVERLAY` names a file in
  the go command's overlay format (`go build -overlay`), whose replacement
  sources every package load reads. gotogether's qweave writes it
  with woven code (Orchestrion aspects), so `//quint:action` records in a
  simulation; any tool can.

- The module path, `github.com/chucklehead-dev/gosim`, so the fork is a
  module others can require. One commit at the top makes it, with this
  script (run again after rebasing onto a newer upstream):

  ```sh
  git grep -l 'jellevandenhooff/gosim' -- ':!FORK.md' |
    xargs perl -pi -e 's#github\.com/jellevandenhooff/gosim#github.com/chucklehead-dev/gosim#g'
  ```
- A Go version check: the hooks match one release's standard library
  (`gosimtool.SupportedGo`, Go 1.27). Translation resolves the effective
  `go env GOVERSION GOROOT`, verifies that root's version, and pins package
  loading and builds to that selection. The version and root also identify
  translation cache entries. Use `GOTOOLCHAIN=go1.27.3`; the module requires
  at least that patch release. Go 1.26 is no longer a supported target.

gotogether uses it through `qgosim`, which requires this module.

## Go 1.27 port acceptance

The port updates hooks for standard library changes, including direct runtime
atomics in `sync.Pool`, deterministic `hash/maphash` support, and timer channel
semantics. This description is a validation checklist, not a claim that every
standard library API is simulated. Missing hooks remain translation errors;
unsupported FIPS entropy timing fails explicitly rather than sampling host time.

From a clean checkout, use the module's minimum toolchain and run:

```sh
export GOTOOLCHAIN=go1.27.3
export GOWORK=off
export GOFLAGS=-trimpath
go run github.com/go-task/task/v3/cmd/task toolchain-check
go run github.com/go-task/task/v3/cmd/task --concurrency 1 check-fmt go-test
```

`go-test` builds the behavior and nemesis simulation binaries, prepares the
self-test binaries, and runs the repository tests, including metatests. Both
steps must pass before treating the non-race port as accepted. CI runs these
checks with the exact version specified by `go.mod`, without automatic
selection of another toolchain.

Race-mode simulation (`-race`) remains unsupported. Its separate acceptance
command is below; it is deliberately not part of the passing non-race CI gate
until the runtime race integration is ported. Keep explicit unsupported-operation
failures until their simulated implementations are validated.

```sh
go run github.com/go-task/task/v3/cmd/task --concurrency 1 go-race-test
```

The encompassing `task test` target still includes race tests and therefore is
not a supported substitute for the non-race acceptance commands above.

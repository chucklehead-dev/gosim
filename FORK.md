# gosim, forked for gotogether

This repository is [gosim](https://github.com/jellevandenhooff/gosim), by
Jelle van den Hooff (MIT, see LICENSE), at ffd3a61 with the changes
[gotogether](https://github.com/chucklehead-dev/gotogether) needs, kept as ordinary commits on top of
upstream's history so that they can be rebased onto a newer upstream:

- Go 1.26 support: hooks for the runtime entry points added or moved since
  Go 1.23 (`internal/hooks/go123/go126.go`), the translator reporting every
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
  (`gosimtool.SupportedGo`, Go 1.26), and `gosim` refuses to run under
  another with the fix (`GOTOOLCHAIN=go1.26.1`).

gotogether uses it through `qgosim`, which requires this module. Race mode
(`-race`) does not build on Go 1.26 yet.

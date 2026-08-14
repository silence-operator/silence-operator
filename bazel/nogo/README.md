# nogo

Runs a subset of `.golangci.yml`'s linters at Bazel build time (`bazel build //...`), via
[rules_go's nogo](https://github.com/bazelbuild/rules_go/blob/master/go/nogo.rst). It benefits
from Bazel's caching: unchanged packages aren't re-analyzed.

nogo only works with checks exported as a `golang.org/x/tools/go/analysis.Analyzer`. Covered here:

- `govet` — the full `golang.org/x/tools/go/analysis/passes` suite (`TOOLS_NOGO`), a superset of
  `go vet`.
- `staticcheck` — the `SA*` bug-detector checks from `honnef.co/go/tools/staticcheck`
  (`staticcheck.bzl`). The `simple`/`stylecheck`/`quickfix` (`S*`/`ST*`/`QF*`) families are style
  suggestions rather than bugs and are skipped here to keep the label list manageable; regenerate
  `staticcheck.bzl` (see the comment in that file) after bumping `honnef.co/go/tools` in `go.mod`.
- `unused` — `honnef.co/go/tools/unused`, via the `//bazel/nogo/unused` wrapper package (its
  `Analyzer` var isn't natively `*analysis.Analyzer`).
- `ineffassign`, `ginkgolinter`, `errcheck` — each ships its own `Analyzer`.

Not covered, because they're golangci-lint-internal implementations with no importable
`Analyzer`: `copyloopvar`, `dupl`, `goconst`, `gocyclo`, `lll`, `misspell`, `nakedret`,
`prealloc`, `revive`, `unconvert`, `unparam`. `golangci-lint` (see the `Lint` GitHub Actions
workflow) remains the source of truth for those and for the full lint config.

`config.json` excludes controller-gen output (`zz_generated.*.go`), mirroring
`.golangci.yml`'s `generated: lax`.

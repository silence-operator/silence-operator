# Development

## BUILD.bazel files

You Don't have to hand-edit `BUILD.bazel` files. After adding, removing, or moving Go files, regenerate them
with Gazelle:

```sh
bazel run //:gazelle
```

## License headers

Don't hand-type license headers either. Each package declares a `license_check()` target
(see `bazel/addlicense.bzl`) that exposes an `addlicense` test and an `addlicense_fix` target.
To stamp missing headers for a package:

```sh
bazel run //cmd:addlicense_fix
```

Gazelle doesn't know about `license_check()` — it only manages `go_library`/`go_test`/`go_binary`.
For a brand-new package, add the `load("//bazel:addlicense.bzl", "license_check")` line, the
`package(default_package_metadata = ["//:package_metadata"])` line, and a `license_check()` call
by hand (copy an existing `BUILD.bazel`), then run `addlicense_fix` to stamp the header comment.

`bazel coverage --config=ci //...` (see [README.md](README.md#test-coverage)) runs the
`addlicense` test as part of the normal build, so a missing/incorrect header fails CI.

## Coverage

```sh
bazel coverage --config=ci //...
genhtml -o coverage-report/report bazel-out/_coverage/_coverage_report.dat
open coverage-report/report/index.html  # xdg-open on Linux
```

`genhtml` ships with `lcov` (`brew install lcov` / `apt-get install lcov`). `coverage-report/` is
gitignored. On lcov 2.x (e.g. Ubuntu 24.04), `genhtml` may reject Bazel's merged `.dat` as
"inconsistent"; add `--ignore-errors inconsistent` if so. See [README.md](README.md#test-coverage)
for how the same command feeds the per-PR coverage comment in CI.

## Linting

You can run golangci-lint as bazel target that pins the same version CI uses:

```sh
bazel run //:lint
```

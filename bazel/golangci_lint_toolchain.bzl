# Copyright 2026 Silence-Operator Maintainers
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Repository rule for downloading a pinned golangci-lint binary.

Kept in sync with the version `.github/workflows/lint.yml`'s golangci-lint-action pins, so
`bazel run //:lint` and CI's Lint job report the same issues.
"""

VERSION = "2.12.2"

_GOLANGCI_LINT_URLS = {
    "linux_amd64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-linux-amd64.tar.gz",
        "8df580d2670fed8fa984aac0507099af8df275e665215f5c7a2ae3943893a553",
    ),
    "linux_arm64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-linux-arm64.tar.gz",
        "44cd40a8c76c86755375adfeea52cfd3533cb43d7bd647771e0ae065e166df3a",
    ),
    "darwin_amd64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-darwin-amd64.tar.gz",
        "f6f06d94b6241521c53d15450c5209b028270bf966f842afb11c030c79f5bc16",
    ),
    "darwin_arm64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.12.2/golangci-lint-2.12.2-darwin-arm64.tar.gz",
        "a9c54498731b3128f79e090be6110f3e5fffccc617b08142ed244d4126c73f29",
    ),
}

def _detect_platform(rctx):
    os_name = rctx.os.name.lower()
    if "linux" in os_name:
        os = "linux"
    elif "mac" in os_name:
        os = "darwin"
    else:
        fail("Unsupported OS: " + os_name)

    arch = rctx.os.arch.lower()
    if arch in ("x86_64", "amd64"):
        arch = "amd64"
    elif arch in ("aarch64", "arm64"):
        arch = "arm64"
    else:
        fail("Unsupported arch: " + arch)

    return os, arch

def _golangci_lint_repo_impl(rctx):
    os, arch = _detect_platform(rctx)
    platform = os + "_" + arch
    entry = _GOLANGCI_LINT_URLS.get(platform)
    if not entry:
        fail("No golangci-lint binary for platform: " + platform)
    url, sha256 = entry

    # Archive extracts into golangci-lint-<version>-<os>-<arch>/; strip it to the repo root.
    rctx.download_and_extract(
        url = url,
        sha256 = sha256,
        strip_prefix = "golangci-lint-" + VERSION + "-" + os + "-" + arch,
    )
    rctx.file("BUILD.bazel", 'exports_files(["golangci-lint", "LICENSE"])\n')

golangci_lint_repo = repository_rule(
    implementation = _golangci_lint_repo_impl,
    doc = "Downloads the pinned v" + VERSION + " golangci-lint release binary for the host platform.",
)

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

VERSION = "2.13.1"

_GOLANGCI_LINT_URLS = {
    "linux_amd64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.13.1/golangci-lint-2.13.1-linux-amd64.tar.gz",
        "b17bfbc9d4aaa48be7f4f1ce3240bc3d8200c870c072bacf15c26219e2cfb9cc",
    ),
    "linux_arm64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.13.1/golangci-lint-2.13.1-linux-arm64.tar.gz",
        "908317c23db18448f924e853b3d8a659fd919614cd438f224810a4053daa2607",
    ),
    "darwin_amd64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.13.1/golangci-lint-2.13.1-darwin-amd64.tar.gz",
        "2c373363953e4e0bee2a03b7fe864a5eb6a3822927cb077d9ca33f2ae3cb2da2",
    ),
    "darwin_arm64": (
        "https://github.com/golangci/golangci-lint/releases/download/v2.13.1/golangci-lint-2.13.1-darwin-arm64.tar.gz",
        "0c9818baf6fb8ad26c6d2ef51b68d5a1e260ef07727036b1431647cc44637c7c",
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

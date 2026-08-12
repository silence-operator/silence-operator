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

"""Repository rule for downloading a pinned helm binary."""

_VERSION = "4.2.3"

_HELM_URLS = {
    "linux_amd64": (
        "https://get.helm.sh/helm-v4.2.3-linux-amd64.tar.gz",
        "e9b88b4ee95b18c706839c28d3a0220e5bc470e9cd9262410c90793c45ff8b7c",
    ),
    "linux_arm64": (
        "https://get.helm.sh/helm-v4.2.3-linux-arm64.tar.gz",
        "21abd9354d39b2cd79a8d76be6912cd137a983cbf997193503fb8a6a6e2f2785",
    ),
    "darwin_amd64": (
        "https://get.helm.sh/helm-v4.2.3-darwin-amd64.tar.gz",
        "ff3ac86755a45f3422473bc1200776aac0fe04c5766abe6ca66699f7b564b23b",
    ),
    "darwin_arm64": (
        "https://get.helm.sh/helm-v4.2.3-darwin-arm64.tar.gz",
        "048ecf5ad3160f83d918f9fe945238d2132b079640f7b106175331c25f242c64",
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

def _helm_repo_impl(rctx):
    os, arch = _detect_platform(rctx)
    platform = os + "_" + arch
    entry = _HELM_URLS.get(platform)
    if not entry:
        fail("No helm binary for platform: " + platform)
    url, sha256 = entry

    # The tarball unpacks to {os}-{arch}/helm -- strip the prefix so the
    # binary lands at the repo root as simply "helm".
    rctx.download_and_extract(
        url = url,
        sha256 = sha256,
        stripPrefix = os + "-" + arch,
    )
    rctx.file("BUILD.bazel", 'exports_files(["helm"])\n')

helm_repo = repository_rule(
    implementation = _helm_repo_impl,
    doc = "Downloads the pinned v" + _VERSION + " helm release binary for the host platform.",
)

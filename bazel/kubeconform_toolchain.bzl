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

"""Repository rule for downloading a pinned kubeconform binary.

Adapted from datavant/rules_gitops (toolchains/kubeconform.bzl), Apache-2.0.
"""

_VERSION = "0.8.0"

_KUBECONFORM_URLS = {
    "linux_amd64": (
        "https://github.com/yannh/kubeconform/releases/download/v0.8.0/kubeconform-linux-amd64.tar.gz",
        "9bc2bffbf71f261128533edaf912153948b7ff238f9a531ae6d34466ec287883",
    ),
    "linux_arm64": (
        "https://github.com/yannh/kubeconform/releases/download/v0.8.0/kubeconform-linux-arm64.tar.gz",
        "1f53fc8e81258197a35e8603054162a5af1de8c5af13746c71ab680d9534ed87",
    ),
    "darwin_amd64": (
        "https://github.com/yannh/kubeconform/releases/download/v0.8.0/kubeconform-darwin-amd64.tar.gz",
        "71dbc87ac9f24099a62b93570e65aa06312ba6ac8aea63b7f86e9d999edf5a92",
    ),
    "darwin_arm64": (
        "https://github.com/yannh/kubeconform/releases/download/v0.8.0/kubeconform-darwin-arm64.tar.gz",
        "f84f4dfbebf4a6b0b230385fa065a39ea35e02608c2b50d025dcf64775a69d67",
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

def _kubeconform_repo_impl(rctx):
    os, arch = _detect_platform(rctx)
    platform = os + "_" + arch
    entry = _KUBECONFORM_URLS.get(platform)
    if not entry:
        fail("No kubeconform binary for platform: " + platform)
    url, sha256 = entry

    rctx.download_and_extract(url = url, sha256 = sha256)
    rctx.file("BUILD.bazel", 'exports_files(["kubeconform"])\n')

kubeconform_repo = repository_rule(
    implementation = _kubeconform_repo_impl,
    doc = "Downloads the pinned v" + _VERSION + " kubeconform release binary for the host platform.",
)

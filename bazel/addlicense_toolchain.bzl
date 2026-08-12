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

"""Repository rule for downloading a pinned addlicense binary."""

_VERSION = "1.2.0"

_ADDLICENSE_URLS = {
    "linux_amd64": (
        "https://github.com/google/addlicense/releases/download/v1.2.0/addlicense_v1.2.0_Linux_x86_64.tar.gz",
        "6924296771234b1ba6c969e109b4459cff75e15b259676b61df874a92734b242",
    ),
    "linux_arm64": (
        "https://github.com/google/addlicense/releases/download/v1.2.0/addlicense_v1.2.0_Linux_arm64.tar.gz",
        "b14dde867a6dbcd41fdf45096831bc0f6500bf0134693d38c56701e179f3f353",
    ),
    "darwin_amd64": (
        "https://github.com/google/addlicense/releases/download/v1.2.0/addlicense_v1.2.0_macOS_x86_64.tar.gz",
        "335b47a28bab66e39494222f711c22b140d14d6f6d0a08ab9ea7c6f315588c83",
    ),
    "darwin_arm64": (
        "https://github.com/google/addlicense/releases/download/v1.2.0/addlicense_v1.2.0_macOS_arm64.tar.gz",
        "0597305c619f7349748e830e38322dbe7d13cf86cf415b289c70f5cecbb4bf63",
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

def _addlicense_repo_impl(rctx):
    os, arch = _detect_platform(rctx)
    platform = os + "_" + arch
    entry = _ADDLICENSE_URLS.get(platform)
    if not entry:
        fail("No addlicense binary for platform: " + platform)
    url, sha256 = entry

    rctx.download_and_extract(url = url, sha256 = sha256)
    rctx.file("BUILD.bazel", 'exports_files(["addlicense"])\n')

addlicense_repo = repository_rule(
    implementation = _addlicense_repo_impl,
    doc = "Downloads the pinned v" + _VERSION + " addlicense release binary for the host platform.",
)

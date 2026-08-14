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

"""Repository rule for downloading a pinned kustomize binary."""

VERSION = "5.8.1"

_KUSTOMIZE_URLS = {
    "linux_amd64": (
        "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.8.1/kustomize_v5.8.1_linux_amd64.tar.gz",
        "029a7f0f4e1932c52a0476cf02a0fd855c0bb85694b82c338fc648dcb53a819d",
    ),
    "linux_arm64": (
        "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.8.1/kustomize_v5.8.1_linux_arm64.tar.gz",
        "0953ea3e476f66d6ddfcd911d750f5167b9365aa9491b2326398e289fef2c142",
    ),
    "darwin_amd64": (
        "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.8.1/kustomize_v5.8.1_darwin_amd64.tar.gz",
        "ee7cf0c1e3592aa7bb66ba82b359933a95e7f2e0b36e5f53ed0a4535b017f2f8",
    ),
    "darwin_arm64": (
        "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.8.1/kustomize_v5.8.1_darwin_arm64.tar.gz",
        "8886f8a78474e608cc81234f729fda188a9767da23e28925802f00ece2bab288",
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

def _kustomize_repo_impl(rctx):
    os, arch = _detect_platform(rctx)
    platform = os + "_" + arch
    entry = _KUSTOMIZE_URLS.get(platform)
    if not entry:
        fail("No kustomize binary for platform: " + platform)
    url, sha256 = entry

    rctx.download_and_extract(url = url, sha256 = sha256)
    rctx.file("BUILD.bazel", 'exports_files(["kustomize"])\n')

kustomize_repo = repository_rule(
    implementation = _kustomize_repo_impl,
    doc = "Downloads the pinned v" + VERSION + " kustomize release binary for the host platform.",
)

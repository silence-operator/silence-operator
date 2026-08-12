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

"""Rule for validating Kubernetes manifests against JSON schemas using kubeconform.

Adapted from datavant/rules_gitops (rules/kubeconform.bzl), Apache-2.0; fixed
the tool rlocation path for the external toolchain repo (see addlicense.bzl).
"""

BASH_RLOCATION_FUNCTION = r"""\
# --- begin runfiles.bash initialization v3 ---
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
  source "$0.runfiles/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  { echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---
"""

script_template = """
$(rlocation {kubeconform_tool}) -kubernetes-version {kubernetes_version} -schema-location "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/{k8s_schema_commit}/{{{{.NormalizedKubernetesVersion}}}}-standalone{{{{.StrictSuffix}}}}/{{{{.ResourceKind}}}}{{{{.KindSuffix}}}}.json" -schema-location "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/{k8s_schema_commit}/{{{{.NormalizedKubernetesVersion}}}}/{{{{.ResourceKind}}}}{{{{.KindSuffix}}}}.json" -schema-location "https://raw.githubusercontent.com/datreeio/CRDs-catalog/{crd_schema_commit}/{{{{.Group}}}}/{{{{.ResourceKind}}}}_{{{{.ResourceAPIVersion}}}}.json" {extra_args} -summary -output json {manifests_path}
"""

def _rlocation_path(ctx, file):
    # An external repo's short_path already starts with "../<repo>/..." -- that's the rlocation path; only main-repo files need workspace_name prepended.
    if file.short_path.startswith("../"):
        return file.short_path[3:]
    return ctx.workspace_name + "/" + file.short_path

def _kubeconform_impl(ctx):
    runfiles_relative_tool_path = _rlocation_path(ctx, ctx.attr.kubeconform_tool[DefaultInfo].files_to_run.executable)

    script_content = script_template.format(
        kubeconform_tool = runfiles_relative_tool_path,
        kubernetes_version = ctx.attr.kubernetes_version,
        k8s_schema_commit = ctx.attr.k8s_schema_commit,
        crd_schema_commit = ctx.attr.crd_schema_commit,
        extra_args = ctx.attr.extra_args,
        manifests_path = " ".join([f.short_path for f in ctx.files.srcs]),
    )

    ctx.actions.write(
        output = ctx.outputs.executable,
        content = BASH_RLOCATION_FUNCTION + script_content,
    )

    # Pre-built binaries from exports_files() don't join default_runfiles the way go_binary does, so add them explicitly.
    runfiles = ctx.runfiles(files = ctx.files.srcs + ctx.files.kubeconform_tool)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

kubeconform_test = rule(
    implementation = _kubeconform_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
            doc = "Manifests to test.",
        ),
        "kubeconform_tool": attr.label(
            default = Label("@kubeconform_toolchain//:kubeconform"),
            executable = True,
            allow_files = True,
            cfg = "exec",
        ),
        "_runfiles_lib": attr.label(
            default = "@bazel_tools//tools/bash/runfiles",
        ),
        "kubernetes_version": attr.string(
            default = "1.36.3",
            doc = "Kubernetes version to validate against.",
        ),
        "k8s_schema_commit": attr.string(
            default = "c8f4e61c63bc529749125ac566bccc6986e08d45",
            doc = "Commit hash of yannh/kubernetes-json-schema to fetch schemas from.",
        ),
        "crd_schema_commit": attr.string(
            default = "52b0261318acc7dd0b66e032759b1f218216b980",
            doc = "Commit hash of datreeio/CRDs-catalog to fetch CRD schemas from.",
        ),
        "extra_args": attr.string(
            default = "",
        ),
    },
    test = True,
)

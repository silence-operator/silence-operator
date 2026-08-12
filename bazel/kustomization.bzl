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

"""Rule for building Kubernetes manifests from kustomize overlays.

Adapted from datavant/rules_gitops (rules/kustomization.bzl), Apache-2.0.
"""

def _kustomization_impl(ctx):
    # Stage srcs at their workspace-relative paths so overlay->base relative refs (../../base) resolve.
    workspace_name = ctx.label.workspace_name
    staging_prefix = workspace_name if workspace_name else "_"

    copy_cmds = [
        "mkdir -p $(dirname $staging/{staged}) && cp {src} $staging/{staged}".format(
            src = f.path,
            staged = f.short_path[3:] if f.short_path.startswith("../") else staging_prefix + "/" + f.short_path,
        )
        for f in ctx.files.srcs
    ]

    # Stage helm chart deps under charts/, adjacent to every kustomization.yaml in srcs.
    kmz_dirs = []
    for f in ctx.files.srcs:
        if f.basename == "kustomization.yaml":
            short = f.short_path
            staged_path = short[3:] if short.startswith("../") else staging_prefix + "/" + short
            last_slash = staged_path.rfind("/")
            kmz_dirs.append(staged_path[:last_slash])

    for f in ctx.files.deps:
        without_dotdot = f.short_path[3:]
        slash = without_dotdot.find("/")
        path_in_repo = without_dotdot[slash + 1:]

        for kdir in kmz_dirs:
            staged = kdir + "/charts/" + path_in_repo
            copy_cmds.append(
                "mkdir -p $(dirname $staging/{staged}) && cp {src} $staging/{staged}".format(
                    src = f.path,
                    staged = staged,
                ),
            )

    kustomize_dir = "$staging/" + staging_prefix + "/" + ctx.label.package

    all_cmds = " && \\\n      ".join(copy_cmds)

    script_content = """
      staging=$(mktemp -d) && \\
      {all_cmds} && \\
      {kustomize} build \\
              --load-restrictor LoadRestrictionsNone \\
              --enable-helm \\
              --helm-command {helm} \\
              {kustomize_directory} > {out_file}
    """.format(
        out_file = ctx.outputs.out.path,
        all_cmds = all_cmds,
        kustomize = ctx.attr.kustomize_tool[DefaultInfo].files_to_run.executable.path,
        helm = ctx.attr.helm_tool[DefaultInfo].files_to_run.executable.path,
        kustomize_directory = kustomize_dir,
    )

    ctx.actions.run_shell(
        outputs = [ctx.outputs.out],
        inputs = ctx.files.srcs + ctx.files.deps,
        tools = [ctx.attr.kustomize_tool[DefaultInfo].files_to_run, ctx.attr.helm_tool[DefaultInfo].files_to_run],
        command = script_content,
    )

    return [DefaultInfo(files = depset([ctx.outputs.out]))]

kustomization = rule(
    implementation = _kustomization_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
            doc = "Manifests to build. All files are staged preserving their workspace-relative paths; kustomize runs on the rule's package directory.",
        ),
        "deps": attr.label_list(
            allow_files = True,
            doc = "External Helm chart archives. Staged under charts/ adjacent to the kustomization directory; set helmGlobals.chartHome: charts in kustomization.yaml.",
        ),
        "kustomize_tool": attr.label(
            default = Label("@kustomize_toolchain//:kustomize"),
            executable = True,
            allow_files = True,
            cfg = "exec",
        ),
        "helm_tool": attr.label(
            default = Label("@helm_toolchain//:helm"),
            executable = True,
            allow_files = True,
            cfg = "exec",
        ),
        "out": attr.output(),
    },
    executable = False,
)

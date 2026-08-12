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

"""Rule for scanning source files for secrets using gitleaks.

Adapted from datavant/rules_gitops (rules/gitleaks.bzl), Apache-2.0. Fixed two
bugs vs. upstream: (1) the tool rlocation path for the external toolchain
repo (see addlicense.bzl), and (2) the scan invocation itself -- upstream
passed a bare "dir <path>" positional that gitleaks' detect subcommand
doesn't accept, so it silently ignored the target and scanned (or, absent a
git repo, no-opped on) the sandbox cwd instead. Fixed with --no-git --source.
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

def _rlocation_path(ctx, file):
    # An external repo's short_path already starts with "../<repo>/..." -- that's the rlocation path; only main-repo files need workspace_name prepended.
    if file.short_path.startswith("../"):
        return file.short_path[3:]
    return ctx.workspace_name + "/" + file.short_path

def _gitleaks_impl(ctx):
    tool_path = _rlocation_path(ctx, ctx.attr.gitleaks_tool[DefaultInfo].files_to_run.executable)
    src_path = _rlocation_path(ctx, ctx.file.src)

    config_arg = ""
    if ctx.file.config:
        config_path = _rlocation_path(ctx, ctx.file.config)
        config_arg = '-c "$(rlocation {config})"'.format(config = config_path)

    baseline_arg = ""
    if ctx.file.baseline:
        baseline_path = _rlocation_path(ctx, ctx.file.baseline)
        baseline_arg = '--baseline-path "$(rlocation {baseline})"'.format(baseline = baseline_path)

    script_content = BASH_RLOCATION_FUNCTION + """
$(rlocation {tool}) detect --no-git --follow-symlinks --max-archive-depth 1 {config_arg} {baseline_arg} --source "$(rlocation {src})"
""".format(
        tool = tool_path,
        config_arg = config_arg,
        baseline_arg = baseline_arg,
        src = src_path,
    )

    ctx.actions.write(
        output = ctx.outputs.executable,
        content = script_content,
    )

    data_files = [ctx.file.src] + ctx.files.gitleaks_tool
    if ctx.file.config:
        data_files.append(ctx.file.config)
    if ctx.file.baseline:
        data_files.append(ctx.file.baseline)

    runfiles = ctx.runfiles(files = data_files)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

gitleaks_test = rule(
    implementation = _gitleaks_impl,
    attrs = {
        "src": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "File, directory, or archive to scan.",
        ),
        "config": attr.label(
            allow_single_file = True,
            doc = "Optional gitleaks.toml config file.",
        ),
        "baseline": attr.label(
            allow_single_file = True,
            doc = "Optional baseline JSON file to suppress known findings (--baseline-path).",
        ),
        "gitleaks_tool": attr.label(
            default = Label("@gitleaks_toolchain//:gitleaks"),
            executable = True,
            allow_files = True,
            cfg = "exec",
        ),
        "_runfiles_lib": attr.label(
            default = "@bazel_tools//tools/bash/runfiles",
        ),
    },
    test = True,
)

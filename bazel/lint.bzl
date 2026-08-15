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

"""Rule for running golangci-lint against the whole module (bazel run //:lint).

Not hermetic and can't be a `bazel test`/`bazel build` target: golangci-lint type-checks via
`go list`, which needs the ambient Go module cache (GOMODCACHE) in the standard layout -- gazelle's
go_deps fetches the same modules into Bazel's own external-repo layout instead, so there's no
cheap way to feed it as sandboxed `data`. CI's Lint job (golangci-lint-action) remains the gate;
this target exists so a local run uses the exact same pinned binary.
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

def _golangci_lint_impl(ctx):
    tool_rloc = _rlocation_path(ctx, ctx.attr.golangci_lint_tool[DefaultInfo].files_to_run.executable)

    script = BASH_RLOCATION_FUNCTION + """
cd "$BUILD_WORKSPACE_DIRECTORY"
$(rlocation {tool}) run "$@"
""".format(tool = tool_rloc)

    ctx.actions.write(output = ctx.outputs.executable, content = script, is_executable = True)

    runfiles = ctx.runfiles(files = ctx.files.golangci_lint_tool)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

golangci_lint = rule(
    implementation = _golangci_lint_impl,
    attrs = {
        "golangci_lint_tool": attr.label(
            default = Label("@golangci_lint_toolchain//:golangci-lint"),
            executable = True,
            allow_files = True,
            cfg = "exec",
        ),
        "_runfiles_lib": attr.label(
            default = "@bazel_tools//tools/bash/runfiles",
        ),
    },
    executable = True,
)

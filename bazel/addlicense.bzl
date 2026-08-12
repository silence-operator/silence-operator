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

"""Rules for checking and stamping license headers using addlicense."""

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

_ADDLICENSE_ATTRS = {
    "srcs": attr.label_list(
        allow_files = True,
        mandatory = True,
    ),
    "license": attr.string(
        mandatory = True,
        doc = "License type passed to addlicense -l (e.g. apache, mit, bsd, mpl).",
    ),
    "copyright": attr.string(
        mandatory = True,
        doc = "Copyright holder passed to addlicense -c (e.g. \"Silence-Operator Maintainers\").",
    ),
    "addlicense_tool": attr.label(
        default = Label("@addlicense_toolchain//:addlicense"),
        executable = True,
        allow_files = True,
        cfg = "exec",
    ),
    "_runfiles_lib": attr.label(
        default = "@bazel_tools//tools/bash/runfiles",
    ),
}

def _addlicense_test_impl(ctx):
    tool_rloc = _rlocation_path(ctx, ctx.attr.addlicense_tool[DefaultInfo].files_to_run.executable)

    file_rloc_paths = "\n".join([
        ctx.workspace_name + "/" + f.short_path
        for f in ctx.files.srcs
    ])

    script = BASH_RLOCATION_FUNCTION + """
FILES=()
while IFS= read -r f; do
  FILES+=("$(rlocation "$f")")
done <<'FILELIST'
{file_rloc_paths}
FILELIST

$(rlocation {tool}) -check -l "{license}" -c "{copyright}" "${{FILES[@]}}"
""".format(
        tool = tool_rloc,
        license = ctx.attr.license,
        copyright = ctx.attr.copyright,
        file_rloc_paths = file_rloc_paths,
    )

    ctx.actions.write(output = ctx.outputs.executable, content = script)

    runfiles = ctx.runfiles(files = ctx.files.srcs + ctx.files.addlicense_tool)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

def _addlicense_impl(ctx):
    tool_rloc = _rlocation_path(ctx, ctx.attr.addlicense_tool[DefaultInfo].files_to_run.executable)

    file_short_paths = "\n".join([f.short_path for f in ctx.files.srcs])

    script = BASH_RLOCATION_FUNCTION + """
FILES=()
while IFS= read -r f; do
  FILES+=("$BUILD_WORKSPACE_DIRECTORY/$f")
done <<'FILELIST'
{file_short_paths}
FILELIST

$(rlocation {tool}) -l "{license}" -c "{copyright}" "${{FILES[@]}}"
""".format(
        tool = tool_rloc,
        license = ctx.attr.license,
        copyright = ctx.attr.copyright,
        file_short_paths = file_short_paths,
    )

    ctx.actions.write(output = ctx.outputs.executable, content = script)

    runfiles = ctx.runfiles(files = ctx.files.addlicense_tool)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

addlicense_test = rule(
    implementation = _addlicense_test_impl,
    attrs = _ADDLICENSE_ATTRS,
    test = True,
)

addlicense = rule(
    implementation = _addlicense_impl,
    attrs = _ADDLICENSE_ATTRS,
    executable = True,
)

def license_check(patterns = ["*.go"], allow_empty = False):
    """Declares :addlicense (test) and :addlicense_fix targets for this package.

    Args:
      patterns: extra glob patterns to check, besides "BUILD.bazel".
      allow_empty: passed through to the underlying glob().
    """
    srcs = native.glob(patterns + ["BUILD.bazel"], allow_empty = allow_empty)
    addlicense_test(
        name = "addlicense",
        size = "small",
        srcs = srcs,
        copyright = "Silence-Operator Maintainers",
        license = "apache",
    )
    addlicense(
        name = "addlicense_fix",
        srcs = srcs,
        copyright = "Silence-Operator Maintainers",
        license = "apache",
    )

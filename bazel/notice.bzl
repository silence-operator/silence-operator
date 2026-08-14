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

"""Rule for regenerating the top-level NOTICE file (bazel run //:notice)."""

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

# Groups go-licenses report rows by exact LicensePath content so identical boilerplate prints once; piped in via single quotes, so kept out of .format() since its braces are awk syntax, not substitution placeholders.
AWK_GROUP_BY_LICENSE_TEXT = r"""
BEGIN { FS = "\t" }
NF < 5 { next }
{
    path = $5
    if (path == "") { next }
    if (!(path in cache)) {
        content = ""
        while ((getline line < path) > 0) content = content line "\n"
        close(path)
        cache[path] = content
    }
    text = cache[path]
    if (!(text in seen)) {
        order[++n] = text
        seen[text] = 1
    }
    members[text] = members[text] sprintf("  %s %s\n", $1, $2)
}
END {
    for (i = 1; i <= n; i++) {
        t = order[i]
        print "================================================================================"
        printf "%s", members[t]
        print "--------------------------------------------------------------------------------"
        printf "%s", t
        print ""
    }
}
"""

def _notice_impl(ctx):
    template_rloc = _rlocation_path(ctx, ctx.file.template)
    go_rloc = _rlocation_path(ctx, ctx.attr.go_tool[DefaultInfo].files_to_run.executable)

    bazel_blocks = []
    for i, (entry, license_file) in enumerate(zip(ctx.attr.bazel_deps, ctx.files.bazel_dep_licenses)):
        name, version, spdx, url = entry.split("|")
        bazel_blocks.append("""
cat <<'NOTICE_ENTRY_{i}'

--------------------------------------------------------------------------------
{name} {version} ({spdx})
{url}
--------------------------------------------------------------------------------
NOTICE_ENTRY_{i}
cat "$(rlocation {license_rloc})"
""".format(
            i = i,
            name = name,
            version = version,
            spdx = spdx,
            url = url,
            license_rloc = _rlocation_path(ctx, license_file),
        ))

    # No pinned release binary exists for go-licenses, so it's invoked via version-pinned "go run" on rules_go's hermetic SDK.
    preamble = """
set -euo pipefail
cd "$BUILD_WORKSPACE_DIRECTORY"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

{{
cat <<'NOTICE_HEADER'
Silence Operator
Copyright 2026 Silence-Operator Maintainers

This product includes software developed by third parties. Their licenses
are reproduced below. Dependencies sharing byte-identical license text are grouped together.

================================================================================
Go module dependencies compiled into the "manager" binary
================================================================================
NOTICE_HEADER

"$(rlocation {go_rloc})" run github.com/google/go-licenses@{go_licenses_version} report {go_targets} --template "$(rlocation {template_rloc})" 2>/dev/null | awk '""".format(
        go_licenses_version = ctx.attr.go_licenses_version,
        go_targets = ctx.attr.go_targets,
        template_rloc = template_rloc,
        go_rloc = go_rloc,
    )

    footer = """'

cat <<'NOTICE_HEADER2'

================================================================================
Build-time tools and Bazel modules (not distributed in the container image)
================================================================================
NOTICE_HEADER2
{bazel_blocks}
}} > "$tmp"

chmod 664 "$tmp"
mv "$tmp" NOTICE
""".format(
        bazel_blocks = "".join(bazel_blocks),
    )

    script = BASH_RLOCATION_FUNCTION + preamble + AWK_GROUP_BY_LICENSE_TEXT + footer

    ctx.actions.write(output = ctx.outputs.executable, content = script, is_executable = True)

    runfiles = ctx.runfiles(files = [ctx.file.template] + ctx.files.bazel_dep_licenses)
    runfiles = runfiles.merge(ctx.attr.go_tool[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.attr._runfiles_lib[DefaultInfo].default_runfiles)

    return [DefaultInfo(executable = ctx.outputs.executable, runfiles = runfiles)]

notice = rule(
    implementation = _notice_impl,
    doc = "Regenerates NOTICE. Needs network on a cold module cache to fetch go-licenses itself; no matching test since staleness detection would be flaky in CI.",
    attrs = {
        "template": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "Go template file passed to `go-licenses report --template`.",
        ),
        "go_tool": attr.label(
            default = Label("@rules_go//go:go"),
            executable = True,
            allow_files = True,
            cfg = "target",
            doc = "rules_go's hermetic host-compatible go binary; must stay in target config, not exec (see go_bin_for_host.bzl).",
        ),
        "go_targets": attr.string(
            default = "./cmd/...",
            doc = "Go package patterns to scan, matching what actually gets compiled into the shipped binary.",
        ),
        "go_licenses_version": attr.string(
            default = "v1.6.0",
            doc = "Pinned github.com/google/go-licenses version passed to `go run`.",
        ),
        "bazel_deps": attr.string_list(
            doc = "One \"name|version|spdx|url\" entry per build-time Bazel module/tool, in the same order as bazel_dep_licenses.",
        ),
        "bazel_dep_licenses": attr.label_list(
            allow_files = True,
            doc = "LICENSE file for each entry in bazel_deps, same order.",
        ),
        "_runfiles_lib": attr.label(
            default = "@bazel_tools//tools/bash/runfiles",
        ),
    },
    executable = True,
)

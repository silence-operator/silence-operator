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

"""Rule for running gazelle, every addlicense_fix, and lint together (bazel run //:tidy).

Not hermetic: it shells out to `bazel` on PATH to invoke the other targets, the same way
bazel/lint.bzl shells out to golangci-lint. Safe to nest -- by the time this script runs, the
outer `bazel run //:tidy` invocation has already finished building and is just executing the
result, so the Bazel server is free to serve the nested calls.
"""

_SCRIPT = """\
set -euo pipefail
cd "$BUILD_WORKSPACE_DIRECTORY"

echo "==> bazel run //:gazelle" >&2
bazel run //:gazelle

for t in $(bazel query 'kind(addlicense, //...)' --output=label 2>/dev/null | grep ':addlicense_fix$'); do
  echo "==> bazel run $t" >&2
  bazel run "$t"
done

echo "==> bazel run //:lint" >&2
bazel run //:lint
"""

def _tidy_impl(ctx):
    ctx.actions.write(output = ctx.outputs.executable, content = _SCRIPT, is_executable = True)
    return [DefaultInfo(executable = ctx.outputs.executable)]

tidy = rule(
    implementation = _tidy_impl,
    executable = True,
)

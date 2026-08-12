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

"""Generated list of honnef.co/go/tools staticcheck SA-check analyzer targets.

Each SA check lives in its own Go package (one per bazel target) so nogo, which
requires one `analysis.Analyzer` per dep, can depend on each individually.

Regenerate after bumping the honnef.co/go/tools version in go.mod:

    bazel query 'kind(go_library, @co_honnef_go_tools//staticcheck/...)' \
        | grep -E "^@co_honnef_go_tools//staticcheck/sa[0-9]+:sa[0-9]+$" | sort -V
"""

STATICCHECK_SA_ANALYZERS = [
    "@co_honnef_go_tools//staticcheck/sa1000:sa1000",
    "@co_honnef_go_tools//staticcheck/sa1001:sa1001",
    "@co_honnef_go_tools//staticcheck/sa1002:sa1002",
    "@co_honnef_go_tools//staticcheck/sa1003:sa1003",
    "@co_honnef_go_tools//staticcheck/sa1004:sa1004",
    "@co_honnef_go_tools//staticcheck/sa1005:sa1005",
    "@co_honnef_go_tools//staticcheck/sa1006:sa1006",
    "@co_honnef_go_tools//staticcheck/sa1007:sa1007",
    "@co_honnef_go_tools//staticcheck/sa1008:sa1008",
    "@co_honnef_go_tools//staticcheck/sa1010:sa1010",
    "@co_honnef_go_tools//staticcheck/sa1011:sa1011",
    "@co_honnef_go_tools//staticcheck/sa1012:sa1012",
    "@co_honnef_go_tools//staticcheck/sa1013:sa1013",
    "@co_honnef_go_tools//staticcheck/sa1014:sa1014",
    "@co_honnef_go_tools//staticcheck/sa1015:sa1015",
    "@co_honnef_go_tools//staticcheck/sa1016:sa1016",
    "@co_honnef_go_tools//staticcheck/sa1017:sa1017",
    "@co_honnef_go_tools//staticcheck/sa1018:sa1018",
    "@co_honnef_go_tools//staticcheck/sa1019:sa1019",
    "@co_honnef_go_tools//staticcheck/sa1020:sa1020",
    "@co_honnef_go_tools//staticcheck/sa1021:sa1021",
    "@co_honnef_go_tools//staticcheck/sa1023:sa1023",
    "@co_honnef_go_tools//staticcheck/sa1024:sa1024",
    "@co_honnef_go_tools//staticcheck/sa1025:sa1025",
    "@co_honnef_go_tools//staticcheck/sa1026:sa1026",
    "@co_honnef_go_tools//staticcheck/sa1027:sa1027",
    "@co_honnef_go_tools//staticcheck/sa1028:sa1028",
    "@co_honnef_go_tools//staticcheck/sa1029:sa1029",
    "@co_honnef_go_tools//staticcheck/sa1030:sa1030",
    "@co_honnef_go_tools//staticcheck/sa1031:sa1031",
    "@co_honnef_go_tools//staticcheck/sa1032:sa1032",
    "@co_honnef_go_tools//staticcheck/sa2000:sa2000",
    "@co_honnef_go_tools//staticcheck/sa2001:sa2001",
    "@co_honnef_go_tools//staticcheck/sa2002:sa2002",
    "@co_honnef_go_tools//staticcheck/sa2003:sa2003",
    "@co_honnef_go_tools//staticcheck/sa3000:sa3000",
    "@co_honnef_go_tools//staticcheck/sa3001:sa3001",
    "@co_honnef_go_tools//staticcheck/sa4000:sa4000",
    "@co_honnef_go_tools//staticcheck/sa4001:sa4001",
    "@co_honnef_go_tools//staticcheck/sa4003:sa4003",
    "@co_honnef_go_tools//staticcheck/sa4004:sa4004",
    "@co_honnef_go_tools//staticcheck/sa4005:sa4005",
    "@co_honnef_go_tools//staticcheck/sa4006:sa4006",
    "@co_honnef_go_tools//staticcheck/sa4008:sa4008",
    "@co_honnef_go_tools//staticcheck/sa4009:sa4009",
    "@co_honnef_go_tools//staticcheck/sa4010:sa4010",
    "@co_honnef_go_tools//staticcheck/sa4011:sa4011",
    "@co_honnef_go_tools//staticcheck/sa4012:sa4012",
    "@co_honnef_go_tools//staticcheck/sa4013:sa4013",
    "@co_honnef_go_tools//staticcheck/sa4014:sa4014",
    "@co_honnef_go_tools//staticcheck/sa4015:sa4015",
    "@co_honnef_go_tools//staticcheck/sa4016:sa4016",
    "@co_honnef_go_tools//staticcheck/sa4017:sa4017",
    "@co_honnef_go_tools//staticcheck/sa4018:sa4018",
    "@co_honnef_go_tools//staticcheck/sa4019:sa4019",
    "@co_honnef_go_tools//staticcheck/sa4020:sa4020",
    "@co_honnef_go_tools//staticcheck/sa4021:sa4021",
    "@co_honnef_go_tools//staticcheck/sa4022:sa4022",
    "@co_honnef_go_tools//staticcheck/sa4023:sa4023",
    "@co_honnef_go_tools//staticcheck/sa4024:sa4024",
    "@co_honnef_go_tools//staticcheck/sa4025:sa4025",
    "@co_honnef_go_tools//staticcheck/sa4026:sa4026",
    "@co_honnef_go_tools//staticcheck/sa4027:sa4027",
    "@co_honnef_go_tools//staticcheck/sa4028:sa4028",
    "@co_honnef_go_tools//staticcheck/sa4029:sa4029",
    "@co_honnef_go_tools//staticcheck/sa4030:sa4030",
    "@co_honnef_go_tools//staticcheck/sa4031:sa4031",
    "@co_honnef_go_tools//staticcheck/sa4032:sa4032",
    "@co_honnef_go_tools//staticcheck/sa5000:sa5000",
    "@co_honnef_go_tools//staticcheck/sa5001:sa5001",
    "@co_honnef_go_tools//staticcheck/sa5002:sa5002",
    "@co_honnef_go_tools//staticcheck/sa5003:sa5003",
    "@co_honnef_go_tools//staticcheck/sa5004:sa5004",
    "@co_honnef_go_tools//staticcheck/sa5005:sa5005",
    "@co_honnef_go_tools//staticcheck/sa5007:sa5007",
    "@co_honnef_go_tools//staticcheck/sa5008:sa5008",
    "@co_honnef_go_tools//staticcheck/sa5009:sa5009",
    "@co_honnef_go_tools//staticcheck/sa5010:sa5010",
    "@co_honnef_go_tools//staticcheck/sa5011:sa5011",
    "@co_honnef_go_tools//staticcheck/sa5012:sa5012",
    "@co_honnef_go_tools//staticcheck/sa6000:sa6000",
    "@co_honnef_go_tools//staticcheck/sa6001:sa6001",
    "@co_honnef_go_tools//staticcheck/sa6002:sa6002",
    "@co_honnef_go_tools//staticcheck/sa6003:sa6003",
    "@co_honnef_go_tools//staticcheck/sa6005:sa6005",
    "@co_honnef_go_tools//staticcheck/sa6006:sa6006",
    "@co_honnef_go_tools//staticcheck/sa9001:sa9001",
    "@co_honnef_go_tools//staticcheck/sa9002:sa9002",
    "@co_honnef_go_tools//staticcheck/sa9003:sa9003",
    "@co_honnef_go_tools//staticcheck/sa9004:sa9004",
    "@co_honnef_go_tools//staticcheck/sa9005:sa9005",
    "@co_honnef_go_tools//staticcheck/sa9006:sa9006",
    "@co_honnef_go_tools//staticcheck/sa9007:sa9007",
    "@co_honnef_go_tools//staticcheck/sa9008:sa9008",
    "@co_honnef_go_tools//staticcheck/sa9009:sa9009",
]

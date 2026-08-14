/*
Copyright 2026 Silence-Operator Maintainers.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package unused re-exports honnef.co/go/tools/unused's analysis.Analyzer for nogo.
package unused

import (
	hcunused "honnef.co/go/tools/unused"
)

// Analyzer is the interface nogo expects: a plain *analysis.Analyzer.
var Analyzer = hcunused.Analyzer.Analyzer

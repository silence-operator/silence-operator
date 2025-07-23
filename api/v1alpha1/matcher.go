/*
Copyright 2025.

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

package v1alpha1

import (
	"fmt"
)

type Matcher struct {
	IsEqual bool   `json:"isEqual,omitempty"`
	IsRegex bool   `json:"isRegex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type Matchers []Matcher

func (m Matchers) String() []string {
	out := make([]string, 0, len(m))

	for _, matcher := range m {
		operator := ""
		if !matcher.IsEqual {
			operator = "!"
		}

		if matcher.IsRegex {
			operator += "~"
		} else {
			operator += "="
		}

		filter := fmt.Sprintf("%s%s%s", matcher.Name, operator, matcher.Value)
		out = append(out, filter)
	}

	return out
}

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

package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"
)

const (
	alertNameLabel = "alertname"
	testAlertName  = "TestAlert"
)

func TestMatchers_String(t *testing.T) {
	tests := []struct {
		name     string
		matchers Matchers
		want     []string
	}{
		{
			name:     "empty",
			matchers: Matchers{},
			want:     []string{},
		},
		{
			name: "equal",
			matchers: Matchers{
				{Name: alertNameLabel, Value: testAlertName, IsEqual: true, IsRegex: false},
			},
			want: []string{"alertname=TestAlert"},
		},
		{
			name: "not equal",
			matchers: Matchers{
				{Name: alertNameLabel, Value: testAlertName, IsEqual: false, IsRegex: false},
			},
			want: []string{"alertname!=TestAlert"},
		},
		{
			name: "regex",
			matchers: Matchers{
				{Name: alertNameLabel, Value: "Test.*", IsEqual: true, IsRegex: true},
			},
			want: []string{"alertname=~Test.*"},
		},
		{
			name: "negative regex",
			matchers: Matchers{
				{Name: alertNameLabel, Value: "Test.*", IsEqual: false, IsRegex: true},
			},
			want: []string{"alertname!~Test.*"},
		},
		{
			name: "multiple matchers preserve order",
			matchers: Matchers{
				{Name: alertNameLabel, Value: testAlertName, IsEqual: true, IsRegex: false},
				{Name: "severity", Value: "critical|warning", IsEqual: true, IsRegex: true},
				{Name: "team", Value: "sre", IsEqual: false, IsRegex: false},
			},
			want: []string{
				"alertname=TestAlert",
				"severity=~critical|warning",
				"team!=sre",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.matchers.String()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matchers.String() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSilence_MarshalPreservesExplicitFalseMatchers(t *testing.T) {
	obj := Silence{Spec: SilenceSpec{
		Comment:  "test",
		Matchers: Matchers{{Name: alertNameLabel, Value: testAlertName, IsEqual: false, IsRegex: false}},
	}}
	obj.Finalizers = []string{SilenceFinalizer}

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got struct {
		Spec struct {
			Matchers []map[string]any `json:"matchers"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got.Spec.Matchers) != 1 {
		t.Fatalf("serialized matchers = %d, want 1", len(got.Spec.Matchers))
	}

	for _, field := range []string{"isEqual", "isRegex"} {
		value, present := got.Spec.Matchers[0][field]
		if !present || value != false {
			t.Errorf("serialized %s = %v (present=%t), want explicit false", field, value, present)
		}
	}
}

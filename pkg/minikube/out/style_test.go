/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package out

import (
	"fmt"
	"strings"
	"testing"
)

func TestApplyPrefix(t *testing.T) {
	tests := []struct {
		prefix, format, expected, description string
	}{
		{
			prefix:      "bar",
			format:      "foo",
			expected:    "barfoo",
			description: "bar prefix",
		},
		{
			prefix:      "",
			format:      "foo",
			expected:    "foo",
			description: "empty prefix",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := applyPrefix(test.prefix, test.format)
			if got != test.expected {
				t.Errorf("Expected %v but got %v", test.expected, got)
			}
		})
	}
}

func TestLowPrefix(t *testing.T) {
	tests := []struct {
		expected    string
		description string
		style       style
	}{
		{
			expected:    lowBullet,
			description: "empty prefix",
		},
		{
			expected:    "bar",
			style:       style{LowPrefix: "bar"},
			description: "lowPrefix",
		},
		{
			expected:    lowBullet,
			style:       style{Prefix: "foo"},
			description: "prefix without spaces",
		},
		{
			expected:    lowIndent,
			style:       style{Prefix: "  foo"},
			description: "prefix with spaces",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := lowPrefix(test.style)
			if got != test.expected {
				t.Errorf("Expected %v but got %v", test.expected, got)
			}
		})
	}
}

func TestApplyStyle(t *testing.T) {
	tests := []struct {
		expected    string
		description string
		styleEnum   StyleEnum
		format      string
		useColor    bool
	}{
		{
			expected:    fmt.Sprintf("%sbar", lowBullet),
			description: "format bar, empty style, color off",
			styleEnum:   Empty,
			useColor:    false,
			format:      "bar",
		},
		{
			expected:    "bar",
			description: "not existing style",
			styleEnum:   9999,
			useColor:    false,
			format:      "bar",
		},
		{
			expected:    fmt.Sprintf("%sfoo", styles[Ready].Prefix),
			description: "format foo, ready style, color on",
			styleEnum:   Ready,
			useColor:    true,
			format:      "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			rawGot := applyStyle(test.styleEnum, test.useColor, test.format)
			got := strings.TrimSpace(rawGot)
			if got != test.expected {
				t.Errorf("Expected '%v' but got '%v'", test.expected, got)
			}
		})
	}
}

func TestApplyTemplateFormating(t *testing.T) {
	tests := []struct {
		expected    string
		description string
		styleEnum   StyleEnum
		format      string
		useColor    bool
		a           []V
	}{
		{
			expected:    fmt.Sprintf("%sbar", lowBullet),
			description: "format bar, empty style, color off",
			styleEnum:   Empty,
			useColor:    false,
			format:      "bar",
		},
		{
			expected:    "bar",
			description: "not existing style",
			styleEnum:   9999,
			useColor:    false,
			format:      "bar",
		},
		{
			expected:    fmt.Sprintf("%sfoo", styles[Ready].Prefix),
			description: "format foo, ready style, color on, a nil",
			styleEnum:   Ready,
			useColor:    true,
			format:      "foo",
		},
		{
			expected:    fmt.Sprintf("%sfoo", styles[Ready].Prefix),
			description: "format foo, ready style, color on",
			styleEnum:   Ready,
			useColor:    true,
			format:      "foo",
		},
		{
			expected:    fmt.Sprintf("%s{{ a }}", styles[Ready].Prefix),
			description: "bad format",
			styleEnum:   Ready,
			useColor:    true,
			format:      "{{ a }}",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			rawGot := ApplyTemplateFormatting(test.styleEnum, test.useColor, test.format, test.a...)
			got := strings.TrimSpace(rawGot)
			if got != test.expected {
				t.Errorf("Expected '%v' but got '%v'", test.expected, got)
			}
		})
	}
}

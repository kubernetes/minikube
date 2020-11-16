/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package translate

import (
	"testing"

	"golang.org/x/text/language"
)

func TestSetPreferredLanguage(t *testing.T) {
	var tests = []struct {
		input string
		want  language.Tag
	}{
		{"", language.AmericanEnglish},
		{"C", language.AmericanEnglish},
		{"zh", language.Chinese},
		{"fr_FR.utf8", language.French},
		{"zzyy.utf8", language.AmericanEnglish},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			// Set something so that we can assert change.
			SetPreferredLanguage(tc.input)

			want, _ := tc.want.Base()
			got, _ := GetPreferredLanguage().Base()
			if got != want {
				t.Errorf("SetPreferredLanguage(%s) = %q, want %q", tc.input, got, want)
			}
		})
	}

}

func TestT(t *testing.T) {
	var tests = []struct {
		description, input, expected string
		langDef, langPref            language.Tag
		translations                 map[string]interface{}
	}{
		{
			description: "empty string not default language",
			input:       "",
			expected:    "",
			langPref:    language.English,
			langDef:     language.Lithuanian,
		},
		{
			description: "empty string and default language",
			input:       "",
			expected:    "",
			langPref:    language.English,
			langDef:     language.English,
		},
		{
			description:  "existing translation",
			input:        "cat",
			expected:     "kot",
			langPref:     language.Lithuanian,
			langDef:      language.English,
			translations: map[string]interface{}{"cat": "kot"},
		},
		{
			description:  "not existing translation",
			input:        "cat",
			expected:     "cat",
			langPref:     language.Lithuanian,
			langDef:      language.English,
			translations: map[string]interface{}{"dog": "pies"},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			defaultLanguage = test.langDef
			preferredLanguage = test.langPref
			Translations = test.translations
			got := T(test.input)
			if test.expected != got {
				t.Errorf("T(%v) should return %v, but got: %v", test.input, test.expected, got)
			}
		})
	}
}

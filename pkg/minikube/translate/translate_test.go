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
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

func TestTranslationFilesValid(t *testing.T) {
	languageFiles, err := filepath.Glob("../../../translations/*.json")
	if err != nil {
		t.Fatalf("failed to get translation files: %v", err)
	}
	for _, filename := range languageFiles {
		lang := filepath.Base(filename)
		t.Run(lang, func(t *testing.T) {
			contents, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("unable to read file %s: %v", filename, err)
			}

			// check if JSON is valid
			if valid := json.Valid(contents); !valid {
				t.Fatalf("%s does not contain valid json", filename)
			}

			// convert file into map
			var entries map[string]string
			if err := json.Unmarshal(contents, &entries); err != nil {
				t.Fatalf("could not unmarshal file %s: %v", filename, err)
			}

			// for each line
			for k, v := range entries {
				// if no translation, skip
				if v == "" {
					continue
				}

				// get all variables (ex. {{.name}})
				keyVariables := distinctVariables(k)
				valueVariables := distinctVariables(v)

				// check if number of original string and translated variables match
				if len(keyVariables) != len(valueVariables) {
					t.Errorf("line %q: %q has mismatching number of variables\noriginal string variables: %s; translated variables: %s", k, v, keyVariables, valueVariables)
					continue
				}

				// for each variable in the original string
				for i, keyVar := range keyVariables {
					// check if translated string has same variable
					if keyVar != valueVariables[i] {
						t.Errorf("line %q: %q has mismatching variables\noriginal string variables: %s do not match translated variables: %s", k, v, keyVariables, valueVariables)
						break
					}
				}
			}
		})
	}
}

func distinctVariables(line string) []string {
	re := regexp.MustCompile(`{{\..+?}}`)

	// get all the variables from the string (possiible duplicates)
	variables := re.FindAllString(line, -1)
	distinctMap := make(map[string]bool)

	// add them to a map to get distinct list of variables
	for _, variable := range variables {
		distinctMap[variable] = true
	}
	distinct := []string{}

	// convert map into slice
	for k := range distinctMap {
		distinct = append(distinct, k)
	}

	// sort the slice to make the comparison easier
	sort.Strings(distinct)

	return distinct
}

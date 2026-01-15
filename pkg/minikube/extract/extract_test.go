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

package extract

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestExtract verifies the end-to-end extraction process.
// It checks if strings are correctly extracted from a sample file and written to a JSON file, matching a predefined expected output.
func TestExtract(t *testing.T) {
	// The file to scan
	paths := []string{"testdata/sample_file.go"}

	// The function we care about
	functions := []string{"extract.PrintToScreen"}

	tempdir := t.TempDir()

	src, err := os.ReadFile("testdata/test.json")
	if err != nil {
		t.Fatalf("Reading json file: %v", err)
	}

	tempfile := filepath.Join(tempdir, "tmpdata.json")
	err = os.WriteFile(tempfile, src, 0666)
	if err != nil {
		t.Fatalf("Writing temp json file: %v", err)
	}

	expected := map[string]interface{}{
		"Hint: This is not a URL, come on.":         "",
		"Holy cow I'm in a loop!":                   "Something else",
		"This is a variable with a string assigned": "",
		"This was a choice: %s":                     "Something",
		"Wow another string: %s":                    "",
	}

	err = TranslatableStrings(paths, functions, tempdir)
	if err != nil {
		t.Fatalf("Error translating strings: %v", err)
	}

	f, err := os.ReadFile(tempfile)
	if err != nil {
		t.Fatalf("Reading resulting json file: %v", err)
	}

	var got map[string]interface{}

	err = json.Unmarshal(f, &got)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %v", err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Translation JSON not equal: expected %v, got %v", expected, got)
	}

}

// TestExtractShouldReturnErrorOnFunctionWithoutPackage verifies input validation for function names.
// It ensures that the extractor rejects function names that are missing a package qualifier, preventing invalid configuration.
func TestExtractShouldReturnErrorOnFunctionWithoutPackage(t *testing.T) {
	expected := errors.New("Initializing: invalid function string missing_package. Needs package name as well")
	funcs := []string{"missing_package"}
	err := TranslatableStrings([]string{}, funcs, "")
	if err == nil || err.Error() != expected.Error() {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

// TestShouldCheckFile verifies the file filtering logic.
// It ensures that the extractor only processes Go source files and ignores test files or other non-Go files, optimizing performance and correctness.
func TestShouldCheckFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"file.go", true},
		{"path/to/file.go", true},
		{"file_test.go", false},
		{"path/to/file_test.go", false},
		{"file.txt", false},
		{"file.json", false},
	}

	for _, tt := range tests {
		got := shouldCheckFile(tt.path)
		if got != tt.want {
			t.Errorf("shouldCheckFile(%q) = %v; want %v", tt.path, got, tt.want)
		}
	}
}

// TestCheckString verifies the string filtering logic that determines which strings should be translated.
// It ensures that URLs, commands, numbers, and excluded patterns are ignored, while valid user-facing strings are captured.
func TestCheckString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"Hello World"`, "Hello World"},
		{`"Short"`, "Short"},
		{`""`, ""},
		{`"a"`, "a"},
		{`"123"`, ""},
		{`"http://google.com"`, ""},
		{`"https://kubernetes.io"`, ""},
		{`"sudo rm -rf /"`, ""},
		{`"{{.error}}"`, ""}, // Excluded
		{`"Use \\n for newlines"`, "Use \\n for newlines"}, // Backslash handling
	}

	for _, tt := range tests {
		got := checkString(tt.input)
		if got != tt.want {
			t.Errorf("checkString(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

// TestExtractAdvice verifies that the extractor correctly identifies and extracts "Advice" fields from the knownIssues map structure.
// This is crucial for localizing error resolution steps (advice) provided to users, ensuring they can understand how to fix issues in their native language.
func TestExtractAdvice(t *testing.T) {
	src := `
package reason

var knownIssues = map[string]KnownIssue{
	"ISSUE_CODE": {
		Advice: "Try restarting your computer",
	},
	"ANOTHER_ISSUE": {
		Advice: fmt.Sprintf("Run %s command", "minikube start"),
	},
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test source: %v", err)
	}

	e, err := newExtractor([]string{})
	if err != nil {
		t.Fatalf("newExtractor failed: %v", err)
	}

	err = extractAdvice(f, e)
	if err != nil {
		t.Fatalf("extractAdvice failed: %v", err)
	}

	expected := []string{
		"Try restarting your computer",
		"minikube start",
	}

	for _, s := range expected {
		if _, ok := e.translations[s]; !ok {
			t.Errorf("Expected translation %q not found in %v", s, e.translations)
		}
	}
}

// TestExtractFlagHelpText verifies that the extractor correctly identifies and extracts usage/help text from command flags.
// This is crucial for localizing the CLI interface, ensuring users can see help messages in their native language when using minikube commands.
func TestExtractFlagHelpText(t *testing.T) {
	src := `
package main

func init() {
	cmd.Flags().StringVarP(&opts.Name, "name", "n", "default", "The name of the cluster")
	cmd.Flags().Bool("force", false, "Force the operation")
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test source: %v", err)
	}

	e, err := newExtractor([]string{})
	if err != nil {
		t.Fatalf("newExtractor failed: %v", err)
	}

	ast.Inspect(f, func(x ast.Node) bool {
		checkNode(x, e)
		return true
	})

	expected := []string{
		"The name of the cluster",
		"Force the operation",
	}

	for _, s := range expected {
		if _, ok := e.translations[s]; !ok {
			t.Errorf("Expected translation %q not found in %v", s, e.translations)
		}
	}
}

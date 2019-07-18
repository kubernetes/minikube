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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtract(t *testing.T) {
	// The file to scan
	paths := []string{"testdata/sample_file.go"}

	// The function we care about
	functions := []string{"extract.PrintToScreen"}

	tempdir, err := ioutil.TempDir("", "temptestdata")
	if err != nil {
		t.Fatalf("Creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempdir)

	src, err := ioutil.ReadFile("testdata/test.json")
	if err != nil {
		t.Fatalf("Reading json file: %v", err)
	}

	tempfile := filepath.Join(tempdir, "tmpdata.json")
	err = ioutil.WriteFile(tempfile, src, 0666)
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

	f, err := ioutil.ReadFile(tempfile)
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

func TestTranslationsUpToDate(t *testing.T) {
	// Move the working dir to where we would run `make extract` from
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current working dir: %v", err)
	}

	err = os.Chdir("../../..")
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Logf("Chdir to cwd failed: %v", err)
		}
	}()

	// The translation file we're going to check
	exampleFile := "translations/fr-FR.json"
	src, err := ioutil.ReadFile(exampleFile)
	if err != nil {
		t.Fatalf("Reading json file: %v", err)
	}

	// Create a temp file to run the extractor on
	tempdir, err := ioutil.TempDir("", "temptestdata")
	if err != nil {
		t.Fatalf("Creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempdir)

	tempfile := filepath.Join(tempdir, "tmpdata.json")
	err = ioutil.WriteFile(tempfile, src, 0666)
	if err != nil {
		t.Fatalf("Writing temp json file: %v", err)
	}

	// Run the extractor exactly how `make extract` would run, but on the temp file
	err = TranslatableStrings([]string{"cmd", "pkg"}, []string{"translate.T"}, tempdir)
	if err != nil {
		t.Fatalf("Error translating strings: %v", err)
	}

	dest, err := ioutil.ReadFile(tempfile)
	if err != nil {
		t.Fatalf("Reading resulting json file: %v", err)
	}

	var got map[string]interface{}
	var want map[string]interface{}

	err = json.Unmarshal(dest, &got)
	if err != nil {
		t.Fatalf("Populating resulting json: %v", err)
	}

	err = json.Unmarshal(src, &want)
	if err != nil {
		t.Fatalf("Populating original json: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Localized string mismatch (-want, +got):\n%s\n\nRun `make extract` to fix.", diff)
	}

}

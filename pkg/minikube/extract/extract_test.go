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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
	defer func() { // clean up tempdir
		err := os.RemoveAll(tempdir)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", tempdir)
		}
	}()

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

func TestExtractShouldReturnErrorOnFunctionWithoutPackage(t *testing.T) {
	expected := errors.New("Initializing: invalid function string missing_package. Needs package name as well")
	funcs := []string{"missing_package"}
	err := TranslatableStrings([]string{}, funcs, "")
	if err == nil || err.Error() != expected.Error() {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

package extract

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestExtract(t *testing.T) {
	// The file to scan
	paths := []string{"testdata/sample_file.go"}

	// The function we care about
	functions := []string{"extract.PrintToScreen"}

	// The directory where the sample translation file is in
	output := "testdata/"

	expected := map[string]interface{}{
		"Hint: This is not a URL, come on.":         "",
		"Holy cow I'm in a loop!":                   "Something else",
		"This is a variable with a string assigned": "",
		"This was a choice: %s":                     "Something",
		"Wow another string: %s":                    "",
	}

	err := TranslatableStrings(paths, functions, output)

	if err != nil {
		t.Fatalf("Error translating strings: %v", err)
	}

	var got map[string]interface{}
	f, err := ioutil.ReadFile("testdata/test.json")
	if err != nil {
		t.Fatalf("Reading json file: %s", err)
	}

	json.Unmarshal(f, &got)

	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Translation JSON not equal: expected %v, got %v", expected, got)
	}

}

/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package crashreport

import (
	"testing"

	"io/ioutil"

	"os"
	"path/filepath"

	"github.com/bugsnag/bugsnag-go"
	"github.com/stretchr/testify/assert"
)

func TestFileIsNotReadWhenNotExisting(t *testing.T) {
	metaData := bugsnag.MetaData{}
	addFile("not existing", &metaData)
	assert.Empty(t, metaData)
}

func TestRead(t *testing.T) {
	metaData := bugsnag.MetaData{}
	content := "foo\nbar\nqix\n"
	fileName := createTempFile(t, content)
	defer os.Remove(fileName)
	addFile(fileName, &metaData)
	assert.Equal(t, "foo\nbar\nqix\n", metaData["logfile"][filepath.Base(fileName)])
}

func createTempFile(t *testing.T, content string) string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return file.Name()
}

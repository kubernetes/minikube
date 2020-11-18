/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package update

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// fsUpdate updates local filesystem repo files according to the given schema and data.
// Returns if the update actually changed anything, and any error occurred.
func fsUpdate(fsRoot string, schema map[string]Item, data interface{}) (changed bool, err error) {
	var mode os.FileMode = 0644
	for path, item := range schema {
		path = filepath.Join(fsRoot, path)
		// if the item's content is already set, give it precedence over any current file content
		var content []byte
		if item.Content == nil {
			info, err := os.Stat(path)
			if err != nil {
				return false, fmt.Errorf("unable to get file content: %w", err)
			}
			mode = info.Mode()
			content, err = ioutil.ReadFile(path)
			if err != nil {
				return false, fmt.Errorf("unable to read file content: %w", err)
			}
			item.Content = content
		}
		if err := item.apply(data); err != nil {
			return false, fmt.Errorf("unable to update file: %w", err)
		}
		if !bytes.Equal(content, item.Content) {
			// make sure path exists
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return false, fmt.Errorf("unable to create directory: %w", err)
			}
			if err := ioutil.WriteFile(path, item.Content, mode); err != nil {
				return false, fmt.Errorf("unable to write file: %w", err)
			}
			changed = true
		}
	}
	return changed, nil
}

// Loadf returns the file content read as byte slice
func Loadf(path string) []byte {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		klog.Fatalf("Unable to load file %s: %v", path, err)
		return nil
	}
	return blob
}

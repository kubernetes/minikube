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
	"io/ioutil"
	"os"
	"path/filepath"
)

// fsUpdate updates local filesystem repo files according to the given schema and data.
// Returns if the update actually changed anything, and any error occurred.
func fsUpdate(fsRoot string, schema map[string]Item, data interface{}) (changed bool, err error) {
	for path, item := range schema {
		path = filepath.Join(fsRoot, path)
		blob, err := ioutil.ReadFile(path)
		if err != nil {
			return false, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		mode := info.Mode()

		item.Content = blob
		chg, err := item.apply(data)
		if err != nil {
			return false, err
		}
		if chg {
			changed = true
		}
		if err := ioutil.WriteFile(path, item.Content, mode); err != nil {
			return false, err
		}
	}
	return changed, nil
}

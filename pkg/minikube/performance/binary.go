/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package performance

import (
	"os"
	"path/filepath"
)

// Binary holds all necessary information
// to call a minikube binary
type Binary struct {
	path string
}

// Binaries returns the type *Binary for each provided path
func Binaries(paths []string) ([]*Binary, error) {
	var binaries []*Binary
	for _, path := range paths {
		b, err := newBinary(path)
		if err != nil {
			return nil, err
		}
		binaries = append(binaries, b)
	}
	return binaries, nil
}

func newBinary(path string) (*Binary, error) {
	var err error
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}
	// make sure binary exists
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &Binary{
		path: path,
	}, nil
}

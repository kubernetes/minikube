//go:build integration

/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

// package parse findmnt command results.
package findmnt

import "encoding/json"

type Filesystem struct {
	Target  string `json:"target"`
	Source  string `json:"source"`
	FSType  string `json:"fstype"`
	Options string `json:"options"`
}

type Result struct {
	Filesystems []Filesystem `json:"filesystems"`
}

// ParseOutput parse findmnt output.
func ParseOutput(output []byte) (*Result, error) {
	r := &Result{}
	if err := json.Unmarshal(output, r); err != nil {
		return nil, err
	}
	return r, nil
}

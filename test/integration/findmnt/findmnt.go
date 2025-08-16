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

import (
	"encoding/json"
	"slices"
)

// Filesystem is a findmnt --json filesystem node. It may include children
// filesystems.
type Filesystem struct {
	Target   string       `json:"target"`
	Source   string       `json:"source"`
	FSType   string       `json:"fstype"`
	Options  string       `json:"options"`
	Children []Filesystem `json:"children,omitempty"`
}

// Result if findmnt --json result.
type Result struct {
	Filesystems []Filesystem `json:"filesystems"`
}

// ParseOutput parse findmnt --json output.
func ParseOutput(output []byte) (*Result, error) {
	r := &Result{}
	if err := json.Unmarshal(output, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Result) Equal(other *Result) bool {
	if r == other {
		return true
	}
	if other == nil {
		return false
	}
	if !slices.EqualFunc(r.Filesystems, other.Filesystems, func(a, b Filesystem) bool {
		return a.Equal(&b)
	}) {
		return false
	}
	return true
}

func (f *Filesystem) Equal(other *Filesystem) bool {
	if f == other {
		return true
	}
	if other == nil {
		return false
	}
	if f.Target != other.Target {
		return false
	}
	if f.Source != other.Source {
		return false
	}
	if f.FSType != other.FSType {
		return false
	}
	if !slices.EqualFunc(f.Children, other.Children, func(a, b Filesystem) bool {
		return a.Equal(&b)
	}) {
		return false
	}
	return true
}

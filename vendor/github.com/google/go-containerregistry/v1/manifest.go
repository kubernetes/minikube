// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"encoding/json"
	"io"

	"github.com/google/go-containerregistry/v1/types"
)

// Manifest represents the OCI image manifest in a structured way.
type Manifest struct {
	SchemaVersion int64             `json:"schemaVersion"`
	MediaType     types.MediaType   `json:"mediaType"`
	Config        Descriptor        `json:"config"`
	Layers        []Descriptor      `json:"layers"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// Descriptor holds a reference from the manifest to one of its constituent elements.
type Descriptor struct {
	MediaType   types.MediaType   `json:"mediaType"`
	Size        int64             `json:"size"`
	Digest      Hash              `json:"digest"`
	URLs        []string          `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ParseManifest parses the io.ReadCloser's contents into a Manifest.
func ParseManifest(r io.ReadCloser) (*Manifest, error) {
	defer r.Close()
	m := Manifest{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

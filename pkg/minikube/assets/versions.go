/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed versions.json
var versionsJSON []byte

var versions map[string]string

func init() {
	if err := json.Unmarshal(versionsJSON, &versions); err != nil {
		panic(fmt.Sprintf("failed to unmarshal embedded versions.json: %v", err))
	}
}

// Version returns the version string for a dependency from versions.json.
func Version(dep string) string {
	if v, ok := versions[dep]; ok {
		return v
	}
	return ""
}

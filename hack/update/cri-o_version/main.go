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



package main

import (
	"strings"
)

func main() {
	crioData, crioErr := detectCrio()
	crictlData, crictlErr := detectCrictl()
	if crictlErr != nil || crioErr != nil {
		return
	}
	// now crio requires crictl which has the same major&minor version with crio
	// so we only update them together when this condition is satisfied
	if strings.HasPrefix(crictlData.Version, "v"+crioData.MMVersion) {
		crioData.updateCrio()
		crictlData.updateCrictl()
	}

}

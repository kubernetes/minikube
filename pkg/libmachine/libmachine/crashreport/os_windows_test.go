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

	"github.com/stretchr/testify/assert"
)

func TestParseVerOutput(t *testing.T) {
	output := `

Microsoft Windows [version 6.3.9600]

`

	assert.Equal(t, "Microsoft Windows [version 6.3.9600]", parseVerOutput(output))
}

func TestParseSystemInfoOutput(t *testing.T) {
	output := `
Host Name:                 DESKTOP-3A5PULA
OS Name:                   Microsoft Windows 10 Enterprise
OS Version:                10.0.10240 N/A Build 10240
OS Manufacturer:           Microsoft Corporation
OS Configuration:          Standalone Workstation
OS Build Type:             Multiprocessor Free
Registered Owner:          Windows User
`

	assert.Equal(t, "10.0.10240 N/A Build 10240", parseSystemInfoOutput(output))
}

func TestParseNonEnglishSystemInfoOutput(t *testing.T) {
	output := `
Ignored:                 ...
Ignored:                 ...
Version du Système:      10.0.10350
`

	assert.Equal(t, "10.0.10350", parseSystemInfoOutput(output))
}

func TestParseInvalidSystemInfoOutput(t *testing.T) {
	output := "Invalid"

	assert.Empty(t, parseSystemInfoOutput(output))
}

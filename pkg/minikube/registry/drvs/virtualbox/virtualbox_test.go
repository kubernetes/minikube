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

package virtualbox

import "testing"

func TestParseVboxVersion(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		major   int
		minor   int
		wantErr bool
	}{
		{"stable 7.2.6", "7.2.6", 7, 2, false},
		{"with build suffix", "7.1.12_Ubuntur169389", 7, 1, false},
		{"trailing newline", "7.2.6\n", 7, 2, false},
		{"leading whitespace", "  6.1.26", 6, 1, false},
		{"two components", "6.1", 6, 1, false},
		{"old 5.2", "5.2.42", 5, 2, false},
		{"empty string", "", 0, 0, true},
		{"no minor component", "7", 0, 0, true},
		{"non-numeric major", "foo.bar.baz", 0, 0, true},
		{"non-numeric minor", "7.beta.0", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, err := parseVboxVersion(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseVboxVersion(%q) err = %v, wantErr = %v", tt.in, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if major != tt.major || minor != tt.minor {
				t.Errorf("parseVboxVersion(%q) = (%d, %d), want (%d, %d)", tt.in, major, minor, tt.major, tt.minor)
			}
		})
	}
}

func TestVboxArm64Policy(t *testing.T) {
	tests := []struct {
		name        string
		major       int
		minor       int
		wantHealthy bool
		wantWarn    bool
	}{
		{"pre-arm64 6.1", 6, 1, false, false},
		{"pre-arm64 7.0", 7, 0, false, false},
		{"first arm64 7.1", 7, 1, true, true},
		{"arm64 7.1.x still warns", 7, 1, true, true},
		{"current stable 7.2", 7, 2, true, false},
		{"future 7.3", 7, 3, true, false},
		{"future major 8.0", 8, 0, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthy, warn := vboxArm64Policy(tt.major, tt.minor)
			if healthy != tt.wantHealthy || warn != tt.wantWarn {
				t.Errorf("vboxArm64Policy(%d, %d) = (%v, %v), want (%v, %v)",
					tt.major, tt.minor, healthy, warn, tt.wantHealthy, tt.wantWarn)
			}
		})
	}
}

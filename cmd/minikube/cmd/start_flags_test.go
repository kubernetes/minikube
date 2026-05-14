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

package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseHostDNSSearch(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single search line",
			content:  "nameserver 8.8.8.8\nsearch corp.example.com\n",
			expected: []string{"corp.example.com"},
		},
		{
			name:     "multiple domains on one line",
			content:  "nameserver 8.8.8.8\nsearch corp.example.com eng.example.com\n",
			expected: []string{"corp.example.com", "eng.example.com"},
		},
		{
			name:     "multiple search lines",
			content:  "search corp.example.com\nnameserver 8.8.8.8\nsearch eng.example.com\n",
			expected: []string{"corp.example.com", "eng.example.com"},
		},
		{
			name:     "no search line",
			content:  "nameserver 8.8.8.8\nnameserver 8.8.4.4\n",
			expected: nil,
		},
		{
			name:     "empty search line",
			content:  "search \nnameserver 8.8.8.8\n",
			expected: nil,
		},
		{
			name:     "search with extra whitespace",
			content:  "  search   corp.example.com   eng.example.com  \n",
			expected: []string{"corp.example.com", "eng.example.com"},
		},
		{
			name:     "comments and search",
			content:  "# This is a comment\nsearch corp.example.com\nnameserver 8.8.8.8\n",
			expected: []string{"corp.example.com"},
		},
		{
			name:     "empty file",
			content:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp resolv.conf
			tmpDir := t.TempDir()
			resolvPath := filepath.Join(tmpDir, "resolv.conf")
			if err := os.WriteFile(resolvPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test resolv.conf: %v", err)
			}

			got := parseHostDNSSearchFrom(resolvPath)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseHostDNSSearchFrom() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseHostDNSSearch_MissingFile(t *testing.T) {
	got := parseHostDNSSearchFrom("/nonexistent/resolv.conf")
	if got != nil {
		t.Errorf("parseHostDNSSearchFrom(missing file) = %v, want nil", got)
	}
}

func TestMergeDNSSearch(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: nil,
		},
		{
			name:     "a only",
			a:        []string{"corp.com", "eng.com"},
			b:        nil,
			expected: []string{"corp.com", "eng.com"},
		},
		{
			name:     "b only",
			a:        nil,
			b:        []string{"corp.com"},
			expected: []string{"corp.com"},
		},
		{
			name:     "no overlap",
			a:        []string{"corp.com"},
			b:        []string{"eng.com"},
			expected: []string{"corp.com", "eng.com"},
		},
		{
			name:     "with duplicates",
			a:        []string{"corp.com", "eng.com"},
			b:        []string{"eng.com", "ops.com"},
			expected: []string{"corp.com", "eng.com", "ops.com"},
		},
		{
			name:     "all duplicates",
			a:        []string{"corp.com", "eng.com"},
			b:        []string{"corp.com", "eng.com"},
			expected: []string{"corp.com", "eng.com"},
		},
		{
			name:     "preserves order from a first",
			a:        []string{"z.com", "a.com"},
			b:        []string{"m.com", "a.com"},
			expected: []string{"z.com", "a.com", "m.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeDNSSearch(tt.a, tt.b)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("mergeDNSSearch(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestWarnDNSSearchOverlap(t *testing.T) {
	// warnDNSSearchOverlap only emits warnings (via out.WarningT), it doesn't return errors.
	// We just verify it doesn't panic for various inputs.
	tests := []struct {
		name          string
		domains       []string
		clusterDomain string
	}{
		{"nil domains", nil, "cluster.local"},
		{"empty domains", []string{}, "cluster.local"},
		{"empty cluster domain", []string{"corp.com"}, ""},
		{"no overlap", []string{"corp.com", "eng.com"}, "cluster.local"},
		{"exact match", []string{"cluster.local"}, "cluster.local"},
		{"subdomain match", []string{"ns1.cluster.local"}, "cluster.local"},
		{"mixed overlap and non-overlap", []string{"corp.com", "svc.cluster.local"}, "cluster.local"},
		{"custom cluster domain overlap", []string{"my.domain"}, "my.domain"},
		{"no false positive on partial match", []string{"notcluster.local"}, "cluster.local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			warnDNSSearchOverlap(tt.domains, tt.clusterDomain)
		})
	}
}

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

package vmnet

import (
	"strings"
	"testing"
)

func TestValidateSemantics_Valid(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
		mask  string
	}{
		// --- valid triples (all three set, all checks pass) ---
		{
			name:  "valid 192.168 range (maintainer example)",
			start: "192.168.200.1", end: "192.168.200.127", mask: "255.255.255.0",
		},
		{
			name:  "valid 10/8 range",
			start: "10.0.0.1", end: "10.0.0.254", mask: "255.0.0.0",
		},
		{
			name:  "valid 172.16/12 range",
			start: "172.16.0.1", end: "172.16.0.254", mask: "255.255.255.0",
		},
		{
			name:  "end below subnet broadcast leaves room for static assignment",
			start: "192.168.1.1", end: "192.168.1.10", mask: "255.255.255.0",
		},
		{
			name:  "valid /24 with end two below broadcast",
			start: "192.168.0.1", end: "192.168.0.253", mask: "255.255.255.0",
		},
		{
			name:  "valid /16",
			start: "172.16.0.1", end: "172.16.255.253", mask: "255.255.0.0",
		},
		{
			name:  "valid /8",
			start: "10.0.0.1", end: "10.255.255.253", mask: "255.0.0.0",
		},
		// --- all empty -> nil (R8: default path stays behavior-identical) ---
		{name: "all empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSemantics(tt.start, tt.end, tt.mask); err != nil {
				t.Errorf("validateSemantics(%q, %q, %q): unexpected error: %v", tt.start, tt.end, tt.mask, err)
			}
		})
	}
}

func TestValidateSemantics_Invalid(t *testing.T) {
	// validateSemantics assumes per-value-valid inputs (flag parsing and
	// `config set` reject malformed values first), so every case below uses
	// individually valid values that only fail a cross-field rule.
	tests := []struct {
		name     string
		start    string
		end      string
		mask     string
		errMatch string
	}{
		// --- all-or-none (R5): a partial set is rejected (moved here from node/vmnet_test.go) ---
		{
			name:     "one of three set (start only)",
			start:    "192.168.1.1",
			errMatch: "set together",
		},
		{
			name:     "one of three set (end only)",
			end:      "192.168.1.10",
			errMatch: "set together",
		},
		{
			name:     "one of three set (mask only)",
			mask:     "255.255.255.0",
			errMatch: "set together",
		},
		{
			name:  "two of three set (start + end)",
			start: "192.168.1.1", end: "192.168.1.10",
			errMatch: "set together",
		},
		{
			name:  "two of three set (start + mask)",
			start: "192.168.1.1", mask: "255.255.255.0",
			errMatch: "set together",
		},
		{
			name: "two of three set (end + mask)",
			end:  "192.168.1.10", mask: "255.255.255.0",
			errMatch: "set together",
		},

		// --- different subnets rejected (R4) ---
		{
			name: "start and end different subnets", start: "192.168.1.1", end: "192.168.2.10", mask: "255.255.255.0",
			errMatch: "same subnet",
		},
		{
			name: "end outside /16 subnet", start: "10.0.0.1", end: "10.1.0.10", mask: "255.255.0.0",
			errMatch: "same subnet",
		},

		// --- ordering: end <= start rejected (OQ3) ---
		{
			name: "end equal to start rejected", start: "192.168.1.5", end: "192.168.1.5", mask: "255.255.255.0",
			errMatch: "greater than",
		},
		{
			name: "end less than start rejected", start: "192.168.1.10", end: "192.168.1.5", mask: "255.255.255.0",
			errMatch: "greater than",
		},

		// --- start = network / broadcast rejected (OQ6) ---
		{
			name: "start is network address rejected", start: "192.168.1.0", end: "192.168.1.10", mask: "255.255.255.0",
			errMatch: "network address",
		},
		{
			name: "start is broadcast address rejected", start: "192.168.1.255", end: "192.168.1.10", mask: "255.255.255.0",
			errMatch: "broadcast address",
		},
		{
			name: "start is 10/8 network address rejected", start: "10.0.0.0", end: "10.0.0.10", mask: "255.0.0.0",
			errMatch: "network address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSemantics(tt.start, tt.end, tt.mask)
			if err == nil {
				t.Fatalf("validateSemantics(%q, %q, %q): expected error, got nil", tt.start, tt.end, tt.mask)
			}
			if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("validateSemantics(%q, %q, %q): error %q does not contain %q", tt.start, tt.end, tt.mask, err.Error(), tt.errMatch)
			}
		})
	}
}

func TestNormalizeAddress_Valid(t *testing.T) {
	tests := []struct {
		val  string
		want string
	}{
		{val: "192.168.1.1", want: "192.168.1.1"},
		{val: "10.0.0.1", want: "10.0.0.1"},
		{val: "172.16.0.1", want: "172.16.0.1"},
	}
	for _, tt := range tests {
		got, err := NormalizeAddress(tt.val)
		if err != nil {
			t.Errorf("NormalizeAddress(%q): unexpected error: %v", tt.val, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeAddress(%q) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestNormalizeAddress_Invalid(t *testing.T) {
	tests := []struct {
		val      string
		errMatch string
	}{
		{val: "8.8.8.8", errMatch: "RFC 1918"},                    // public
		{val: "100.64.0.1", errMatch: "RFC 1918"},                 // CGN (not IsPrivate)
		{val: "127.0.0.1", errMatch: "RFC 1918"},                  // loopback
		{val: "::1", errMatch: "not a valid IPv4"},                // IPv6
		{val: "2001:db8::1", errMatch: "not a valid IPv4"},        // IPv6
		{val: "::ffff:192.168.1.1", errMatch: "not a valid IPv4"}, // IPv4-mapped is not a plain IPv4
		{val: "not-an-ip", errMatch: "not a valid IPv4"},          // malformed
		{val: "192.168.1.256", errMatch: "not a valid IPv4"},      // bad octet
		{val: "192.168.1", errMatch: "not a valid IPv4"},          // missing octet
	}
	for _, tt := range tests {
		_, err := NormalizeAddress(tt.val)
		if err == nil {
			t.Errorf("NormalizeAddress(%q): expected error, got nil", tt.val)
			continue
		}
		if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
			t.Errorf("NormalizeAddress(%q): error %q does not contain %q", tt.val, err.Error(), tt.errMatch)
		}
	}
}

func TestNormalizeSubnetMask_Valid(t *testing.T) {
	tests := []struct {
		val  string
		want string
	}{
		{val: "255.255.255.0", want: "255.255.255.0"},     // /24
		{val: "255.255.0.0", want: "255.255.0.0"},         // /16
		{val: "255.0.0.0", want: "255.0.0.0"},             // /8
		{val: "255.255.255.254", want: "255.255.255.254"}, // /31
		{val: "255.255.255.255", want: "255.255.255.255"}, // /32
		{val: "0.0.0.0", want: "0.0.0.0"},                 // /0 is contiguous (Size returns 0,32); degenerate but passes the contiguity predicate
	}
	for _, tt := range tests {
		got, err := NormalizeSubnetMask(tt.val)
		if err != nil {
			t.Errorf("NormalizeSubnetMask(%q): unexpected error: %v", tt.val, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeSubnetMask(%q) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestNormalizeSubnetMask_Invalid(t *testing.T) {
	tests := []struct {
		val      string
		errMatch string
	}{
		{val: "255.255.255.1", errMatch: "contiguous"},               // non-contiguous
		{val: "255.0.255.0", errMatch: "contiguous"},                 // non-contiguous
		{val: "::1", errMatch: "not a valid IPv4"},                   // IPv6
		{val: "ffff:ffff:ffff:ffff::", errMatch: "not a valid IPv4"}, // IPv6
		{val: "255.255.255.256", errMatch: "not a valid IPv4"},       // bad octet
		{val: "not-a-mask", errMatch: "not a valid IPv4"},            // malformed
	}
	for _, tt := range tests {
		_, err := NormalizeSubnetMask(tt.val)
		if err == nil {
			t.Errorf("NormalizeSubnetMask(%q): expected error, got nil", tt.val)
			continue
		}
		if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
			t.Errorf("NormalizeSubnetMask(%q): error %q does not contain %q", tt.val, err.Error(), tt.errMatch)
		}
	}
}

func TestIsValidVmnetAddress_Valid(t *testing.T) {
	vals := []string{
		"",            // empty ok (in-progress config sequence)
		"192.168.1.1", // valid RFC 1918
		"10.0.0.1",    // valid 10/8
		"172.16.0.1",  // valid 172.16/12
	}
	for _, val := range vals {
		if err := IsValidVmnetAddress("vmnet-start-address", val); err != nil {
			t.Errorf("IsValidVmnetAddress(%q): unexpected error: %v", val, err)
		}
	}
}

func TestIsValidVmnetAddress_Invalid(t *testing.T) {
	tests := []struct {
		val      string
		errMatch string
	}{
		{val: "8.8.8.8", errMatch: "RFC 1918"},               // public
		{val: "100.64.0.1", errMatch: "RFC 1918"},            // CGN (not IsPrivate)
		{val: "::1", errMatch: "not a valid IPv4"},           // IPv6
		{val: "2001:db8::1", errMatch: "not a valid IPv4"},   // IPv6
		{val: "not-an-ip", errMatch: "not a valid IPv4"},     // malformed
		{val: "192.168.1.256", errMatch: "not a valid IPv4"}, // bad octet
	}
	for _, tt := range tests {
		err := IsValidVmnetAddress("vmnet-start-address", tt.val)
		if err == nil {
			t.Errorf("IsValidVmnetAddress(%q): expected error, got nil", tt.val)
			continue
		}
		if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
			t.Errorf("IsValidVmnetAddress(%q): error %q does not contain %q", tt.val, err.Error(), tt.errMatch)
		}
	}
}

func TestIsValidVmnetSubnetMask_Valid(t *testing.T) {
	vals := []string{
		"",                // empty ok
		"255.255.255.0",   // valid /24
		"255.255.0.0",     // valid /16
		"255.0.0.0",       // valid /8
		"255.255.255.254", // valid /31
		"255.255.255.255", // valid /32
		"0.0.0.0",         // /0 is contiguous (Size returns 0,32); degenerate but passes the contiguity predicate
	}
	for _, val := range vals {
		if err := IsValidVmnetSubnetMask("vmnet-subnet-mask", val); err != nil {
			t.Errorf("IsValidVmnetSubnetMask(%q): unexpected error: %v", val, err)
		}
	}
}

func TestIsValidVmnetSubnetMask_Invalid(t *testing.T) {
	tests := []struct {
		val      string
		errMatch string
	}{
		{val: "255.255.255.1", errMatch: "contiguous"},         // non-contiguous
		{val: "255.0.255.0", errMatch: "contiguous"},           // non-contiguous
		{val: "::1", errMatch: "not a valid IPv4"},             // IPv6
		{val: "255.255.255.256", errMatch: "not a valid IPv4"}, // bad octet
		{val: "not-a-mask", errMatch: "not a valid IPv4"},      // malformed
	}
	for _, tt := range tests {
		err := IsValidVmnetSubnetMask("vmnet-subnet-mask", tt.val)
		if err == nil {
			t.Errorf("IsValidVmnetSubnetMask(%q): expected error, got nil", tt.val)
			continue
		}
		if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
			t.Errorf("IsValidVmnetSubnetMask(%q): error %q does not contain %q", tt.val, err.Error(), tt.errMatch)
		}
	}
}

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

func TestValidateOptions_Valid(t *testing.T) {
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
			if err := ValidateOptions(tt.start, tt.end, tt.mask); err != nil {
				t.Errorf("ValidateOptions(%q, %q, %q): unexpected error: %v", tt.start, tt.end, tt.mask, err)
			}
		})
	}
}

func TestValidateOptions_Invalid(t *testing.T) {
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
		{
			name:     "partial set with invalid value still reports all-or-none first",
			start:    "8.8.8.8",
			errMatch: "set together",
		},

		// --- IPv6 / non-IPv4 rejected (R3) ---
		{
			name: "IPv6 start rejected", start: "::1", end: "192.168.1.10", mask: "255.255.255.0",
			errMatch: "not a valid IPv4",
		},
		{
			name: "IPv6 end rejected", start: "192.168.1.1", end: "2001:db8::1", mask: "255.255.255.0",
			errMatch: "not a valid IPv4",
		},
		{
			name: "IPv6 mask rejected", start: "192.168.1.1", end: "192.168.1.10", mask: "ffff:ffff:ffff:ffff::",
			errMatch: "not a valid IPv4",
		},

		// --- public / non-RFC-1918 IPv4 rejected (R3, OQ2) ---
		{
			name: "public start rejected", start: "8.8.8.8", end: "192.168.1.10", mask: "255.255.255.0",
			errMatch: "RFC 1918",
		},
		{
			name: "public end rejected", start: "192.168.1.1", end: "1.1.1.1", mask: "255.255.255.0",
			errMatch: "RFC 1918",
		},
		{
			name: "CGN 100.64 rejected (not IsPrivate)", start: "100.64.0.1", end: "100.64.0.10", mask: "255.255.255.0",
			errMatch: "RFC 1918",
		},
		{
			name: "loopback 127 rejected", start: "127.0.0.1", end: "127.0.0.10", mask: "255.255.255.0",
			errMatch: "RFC 1918",
		},

		// --- non-contiguous mask rejected (OQ4) ---
		{
			name: "non-contiguous mask .1 rejected", start: "192.168.1.1", end: "192.168.1.10", mask: "255.255.255.1",
			errMatch: "contiguous",
		},
		{
			name: "non-contiguous mask 255.0.255.0 rejected", start: "192.168.1.1", end: "192.168.1.10", mask: "255.0.255.0",
			errMatch: "contiguous",
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

		// --- malformed inputs ---
		{
			name: "malformed start", start: "192.168.1", end: "192.168.1.10", mask: "255.255.255.0",
			errMatch: "not a valid IPv4",
		},
		{
			name: "malformed mask octet", start: "192.168.1.1", end: "192.168.1.10", mask: "255.255.255.256",
			errMatch: "not a valid IPv4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptions(tt.start, tt.end, tt.mask)
			if err == nil {
				t.Fatalf("ValidateOptions(%q, %q, %q): expected error, got nil", tt.start, tt.end, tt.mask)
			}
			if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("ValidateOptions(%q, %q, %q): error %q does not contain %q", tt.start, tt.end, tt.mask, err.Error(), tt.errMatch)
			}
		})
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

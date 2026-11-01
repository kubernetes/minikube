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

// This file has no //go:build tag on purpose: it compiles into both the darwin
// build (vmnet.go) and the non-darwin build (vmnet_stub.go), mirroring the
// tagless vmnet_error.go. The validator is pure stdlib (net/netip + fmt) and is
// imported cross-platform by cmd/minikube/cmd/start_flags.go and
// cmd/minikube/cmd/config/validations.go, so gating it behind //go:build darwin
// would break the Linux build of those importers.

package vmnet

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
)

// rfc1918Prefixes holds the three RFC 1918 private IPv4 address ranges. vmnet
// networks are constrained to RFC 1918 by vmnet-helper, so non-private
// addresses are rejected early. netip.Addr.IsPrivate() is deliberately not
// used: it also matches CGN (100.64/10), loopback, link-local and IPv6 ULA,
// none of which are valid vmnet addresses.
var rfc1918Prefixes = mustParsePrefixes("10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16")

func mustParsePrefixes(cidrs ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, len(cidrs))
	for i, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			// Well-formed literals; unreachable.
			panic(err)
		}
		prefixes[i] = prefix
	}
	return prefixes
}

func isRFC1918(addr netip.Addr) bool {
	for _, p := range rfc1918Prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// NormalizeAddress parses and validates a single vmnet start/end address,
// returning it as a netip.Addr so the value is parsed once at the flag/config
// boundary and every consumer downstream uses the typed value. It enforces the
// per-value rules only (R3): a valid IPv4 address inside an RFC 1918 private
// range. The all-or-none and cross-field rules belong to the driver boundary
// and live in Helper.Validate (see vmnet.go).
func NormalizeAddress(val string) (netip.Addr, error) {
	return parseVmnetAddress(val, "vmnet address")
}

// NormalizeSubnetMask parses and validates a single vmnet subnet mask,
// returning it as a netip.Addr. It enforces the per-value rules only (R3,
// OQ4): a valid IPv4 address that is a contiguous netmask.
func NormalizeSubnetMask(val string) (netip.Addr, error) {
	return parseVmnetMask(val)
}

// validateSemantics validates the three vmnet-helper options together: the
// all-or-none gate (R5) and the cross-field same-subnet / network /
// broadcast / ordering checks (R4, OQ3, OQ6). It backs Helper.Validate, which
// the vfkit and krunkit drivers call from Start() — every path that starts a
// VM validates at the point of use, so no command needs its own call.
//
// The zero netip.Addr marks an unset option: an all-empty triple
// short-circuits to nil (R8: the default path stays behavior-identical).
// Per-value checks are NOT repeated here: flag parsing and `config set`
// normalize each value (NormalizeAddress / NormalizeSubnetMask) before it can
// be persisted.
func validateSemantics(start, end, mask netip.Addr) error {
	// All-or-none (R5). The three options must be set together; an all-empty
	// triple is valid (R8, the default path), but a partial triple is not. The
	// message is context-neutral since it is surfaced by the driver, whatever
	// command led there.
	if n := setCount(start, end, mask); n != 0 && n != 3 {
		return errors.New("vmnet-start-address, vmnet-end-address, and vmnet-subnet-mask must be set together")
	}

	// Cross-field checks require all three values (R4, OQ3, OQ6).
	if !start.IsValid() || !end.IsValid() || !mask.IsValid() {
		return nil
	}

	sip := ipFrom4(start.As4())
	eip := ipFrom4(end.As4())
	m := net.IPMask(ipFrom4(mask.As4()))

	// Same subnet (R4): start and end must lie in the subnet defined by mask.
	if !sip.Mask(m).Equal(eip.Mask(m)) {
		return fmt.Errorf("vmnet start address %q and end address %q are not in the same subnet defined by mask %q", start, end, mask)
	}

	// Usable start (OQ6): start is the gateway and first DHCP address, so it must
	// not be the subnet's network or broadcast address. This is checked before
	// the ordering check (OQ3) because it is a property of start+mask alone: a
	// start equal to the broadcast makes ANY end fail the ordering check, so
	// reporting the broadcast here gives the clearer message (R7).
	network := sip.Mask(m)
	broadcast := broadcastAddress(network, m)
	if sip.Equal(network) {
		return fmt.Errorf("vmnet start address %q is the network address of the subnet", start)
	}
	if sip.Equal(broadcast) {
		return fmt.Errorf("vmnet start address %q is the broadcast address of the subnet", start)
	}

	// Ordering (OQ3): end must be greater than start so the DHCP pool
	// (start+1..end) has at least one address.
	if toUint32(eip) <= toUint32(sip) {
		return fmt.Errorf("vmnet end address %q must be greater than start address %q", end, start)
	}

	return nil
}

// setCount returns how many of the given vmnet options are set (a valid
// netip.Addr). It backs the all-or-none gate in validateSemantics (R5): a
// count of 0 (the default all-empty config) or 3 is valid; anything in
// between is a partial set.
func setCount(vals ...netip.Addr) int {
	n := 0
	for _, v := range vals {
		if v.IsValid() {
			n++
		}
	}
	return n
}

// parseVmnetAddress runs the per-value checks for a start/end address: valid
// IPv4 (parsed with netip, so non-canonical spellings are rejected) and within
// an RFC 1918 private range (R3).
func parseVmnetAddress(val, label string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(val)
	if err != nil || !addr.Is4() {
		return netip.Addr{}, fmt.Errorf("%s %q is not a valid IPv4 address", label, val)
	}
	if !isRFC1918(addr) {
		return netip.Addr{}, fmt.Errorf("%s %q is not in the RFC 1918 private range (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)", label, val)
	}
	return addr, nil
}

// parseVmnetMask runs the per-value checks for the subnet mask: valid IPv4 and
// a contiguous netmask (R3, OQ4). A non-contiguous mask (e.g. 255.255.255.1)
// is rejected because net.IPMask.Size reports it with bits != 32.
func parseVmnetMask(val string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(val)
	if err != nil || !addr.Is4() {
		return netip.Addr{}, fmt.Errorf("vmnet subnet mask %q is not a valid IPv4 address", val)
	}
	m := net.IPMask(ipFrom4(addr.As4()))
	if _, bits := m.Size(); bits != IPv4Bits {
		return netip.Addr{}, fmt.Errorf("vmnet subnet mask %q is not a valid contiguous netmask", val)
	}
	return addr, nil
}

// IsValidVmnetAddress is a per-value validator for the `config set` path,
// which can see only one key's value at a time. It runs only the per-value
// IPv4 + RFC 1918 checks (no cross-field); the same-subnet, ordering and
// all-or-none checks are enforced by the driver at Start() via Helper.Validate.
// The value cannot be empty: there is no reason to store an empty address in
// the config (`minikube config unset` clears a key).
func IsValidVmnetAddress(_, val string) error {
	_, err := parseVmnetAddress(val, "vmnet address")
	return err
}

// IsValidVmnetSubnetMask is a per-value validator for the `config set` path,
// running only the per-value IPv4 + contiguous-mask checks. The value cannot
// be empty; use `minikube config unset` to clear the key.
func IsValidVmnetSubnetMask(_, val string) error {
	_, err := parseVmnetMask(val)
	return err
}

// IPv4Bits is the number of bits in an IPv4 address.
const IPv4Bits = 32

// ipFrom4 converts the 4-byte form of a netip address for the net.IP-based
// mask/ordering math in validateSemantics.
func ipFrom4(b [4]byte) net.IP {
	return net.IP{b[0], b[1], b[2], b[3]}
}

func toUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// broadcastAddress returns the broadcast address of the subnet whose network
// address is `network` and whose mask is `mask`.
func broadcastAddress(network net.IP, mask net.IPMask) net.IP {
	m := net.IP(mask)
	bc := make(net.IP, len(network))
	for i := range network {
		bc[i] = network[i] | ^m[i]
	}
	return bc
}

// AddrPtr returns a pointer to a copy of addr, or nil when addr is unset (the
// zero netip.Addr). ClusterConfig stores the vmnet options as *netip.Addr so
// an unset option marshals to JSON null instead of failing on the zero value.
func AddrPtr(addr netip.Addr) *netip.Addr {
	if !addr.IsValid() {
		return nil
	}
	return &addr
}

// AddrOrZero dereferences an optional config address, mapping nil and the
// zero Addr to the zero Addr (unset) so Helper consumers never see a string.
func AddrOrZero(addr *netip.Addr) netip.Addr {
	if addr == nil {
		return netip.Addr{}
	}
	return *addr
}

// AddrString renders an optional config address, mapping unset to "".
func AddrString(addr *netip.Addr) string {
	if addr == nil || !addr.IsValid() {
		return ""
	}
	return addr.String()
}

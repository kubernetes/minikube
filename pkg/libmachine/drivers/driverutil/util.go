package driverutil

import "strings"

// SplitPortProto splits a string in the format port/protocol, defaulting
// protocol to "tcp" if not provided.
func SplitPortProto(raw string) (port string, protocol string) {
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) == 1 {
		return parts[0], "tcp"
	}
	return parts[0], parts[1]
}

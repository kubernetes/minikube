package driverutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPortProtocol(t *testing.T) {
	tests := []struct {
		raw           string
		expectedPort  string
		expectedProto string
	}{
		{"8080/tcp", "8080", "tcp"},
		{"90/udp", "90", "udp"},
		{"80", "80", "tcp"},
	}

	for _, tc := range tests {
		port, proto := SplitPortProto(tc.raw)
		assert.Equal(t, tc.expectedPort, port)
		assert.Equal(t, tc.expectedProto, proto)
	}
}

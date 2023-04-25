package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/stretchr/testify/assert"
)

func TestParseSecurityRuleProtocol(t *testing.T) {
	tests := []struct {
		raw           string
		expectedProto network.SecurityRuleProtocol
		expectedErr   bool
	}{
		{"tcp", network.TCP, false},
		{"udp", network.UDP, false},
		{"*", network.Asterisk, false},
		{"Invalid", "", true},
	}

	for _, tc := range tests {
		proto, err := parseSecurityRuleProtocol(tc.raw)
		assert.Equal(t, tc.expectedProto, proto)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

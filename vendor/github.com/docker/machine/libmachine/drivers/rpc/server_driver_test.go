package rpcdriver

import (
	"errors"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/stretchr/testify/assert"
)

type panicDriver struct {
	*fakedriver.Driver
	panicErr  error
	returnErr error
}

type FakeStacker struct {
	trace []byte
}

func (fs *FakeStacker) Stack() []byte {
	return fs.trace
}

func (p *panicDriver) Create() error {
	if p.panicErr != nil {
		panic(p.panicErr)
	}
	return p.returnErr
}

func TestRPCServerDriverCreate(t *testing.T) {
	testCases := []struct {
		description  string
		expectedErr  error
		serverDriver *RPCServerDriver
		stacker      Stacker
	}{
		{
			description: "Happy path",
			expectedErr: nil,
			serverDriver: &RPCServerDriver{
				ActualDriver: &panicDriver{
					returnErr: nil,
				},
			},
		},
		{
			description: "Normal error, no panic",
			expectedErr: errors.New("API not available"),
			serverDriver: &RPCServerDriver{
				ActualDriver: &panicDriver{
					returnErr: errors.New("API not available"),
				},
			},
		},
		{
			description: "Panic happened during create",
			expectedErr: errors.New("Panic in the driver: index out of range\nSTACK TRACE"),
			serverDriver: &RPCServerDriver{
				ActualDriver: &panicDriver{
					panicErr: errors.New("index out of range"),
				},
			},
			stacker: &FakeStacker{
				trace: []byte("STACK TRACE"),
			},
		},
	}

	for _, tc := range testCases {
		stdStacker = tc.stacker
		assert.Equal(t, tc.expectedErr, tc.serverDriver.Create(nil, nil))
	}
}

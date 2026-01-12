/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
)

func TestIP(t *testing.T) {
	cases := []struct {
		baseDriver  *BaseDriver
		expectedIP  string
		expectedErr error
	}{
		{&BaseDriver{}, "", errors.New("IP address is not set")},
		{&BaseDriver{IPAddress: "2001:4860:0:2001::68"}, "2001:4860:0:2001::68", nil},
		{&BaseDriver{IPAddress: "192.168.0.1"}, "192.168.0.1", nil},
		{&BaseDriver{IPAddress: "::1"}, "::1", nil},
		{&BaseDriver{IPAddress: "hostname"}, "hostname", nil},
	}

	for _, c := range cases {
		ip, err := c.baseDriver.GetIP()
		assert.Equal(t, c.expectedIP, ip)
		assert.Equal(t, c.expectedErr, err)
	}
}

func TestEngineInstallUrlFlagEmpty(t *testing.T) {
	assert.False(t, EngineInstallURLFlagSet(&CheckDriverOptions{}))
}

func createDriverOptionWithEngineInstall(url string) *CheckDriverOptions {
	return &CheckDriverOptions{
		FlagsValues: map[string]interface{}{"engine-install-url": url},
		CreateFlags: []mcnflag.Flag{mcnflag.StringFlag{Name: "engine-install-url", Value: ""}},
	}
}

func TestEngineInstallUrlFlagDefault(t *testing.T) {
	options := createDriverOptionWithEngineInstall(DefaultEngineInstallURL)
	assert.False(t, EngineInstallURLFlagSet(options))
}

func TestEngineInstallUrlFlagSet(t *testing.T) {
	options := createDriverOptionWithEngineInstall("https://test.docker.com")
	assert.True(t, EngineInstallURLFlagSet(options))
}

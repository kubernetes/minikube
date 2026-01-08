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

package check

import (
	"errors"
	"testing"

	"crypto/tls"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
)

type FakeValidateCertificate struct {
	IsValid bool
	Err     error
}

type FakeCertGenerator struct {
	fakeValidateCertificate *FakeValidateCertificate
}

func (fcg FakeCertGenerator) GenerateCACertificate(certFile, keyFile, org string, bits int) error {
	return nil
}

func (fcg FakeCertGenerator) GenerateCert(opts *cert.Options) error {
	return nil
}

func (fcg FakeCertGenerator) ValidateCertificate(addr string, authOptions *auth.Options) (bool, error) {
	return fcg.fakeValidateCertificate.IsValid, fcg.fakeValidateCertificate.Err
}

func (fcg FakeCertGenerator) ReadTLSConfig(addr string, authOptions *auth.Options) (*tls.Config, error) {
	return nil, nil
}

func TestCheckCert(t *testing.T) {
	errCertsExpired := errors.New("Certs have expired")

	cases := []struct {
		hostURL     string
		authOptions *auth.Options
		valid       bool
		checkErr    error
		expectedErr error
	}{
		{"192.168.99.100:2376", &auth.Options{}, true, nil, nil},
		{"192.168.99.100:2376", &auth.Options{}, false, nil, ErrCertInvalid{wrappedErr: nil, hostURL: "192.168.99.100:2376"}},
		{"192.168.99.100:2376", &auth.Options{}, false, errCertsExpired, ErrCertInvalid{wrappedErr: errCertsExpired, hostURL: "192.168.99.100:2376"}},
	}

	for _, c := range cases {
		fcg := FakeCertGenerator{fakeValidateCertificate: &FakeValidateCertificate{c.valid, c.checkErr}}
		cert.SetCertGenerator(fcg)
		err := checkCert(c.hostURL, c.authOptions)
		assert.Equal(t, c.expectedErr, err)
	}
}

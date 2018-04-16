// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/go-containerregistry/authn"
	"github.com/google/go-containerregistry/name"
)

type bearerTransport struct {
	// Wrapped by bearerTransport.
	inner http.RoundTripper
	// Basic credentials that we exchange for bearer tokens.
	basic authn.Authenticator
	// Holds the bearer response from the token service.
	bearer *authn.Bearer
	// Registry to which we send bearer tokens.
	registry name.Registry
	// See https://tools.ietf.org/html/rfc6750#section-3
	realm string
	// See https://docs.docker.com/registry/spec/auth/token/
	service string
	scope   string
}

var _ http.RoundTripper = (*bearerTransport)(nil)

// RoundTrip implements http.RoundTripper
func (bt *bearerTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	hdr, err := bt.bearer.Authorization()
	if err != nil {
		return nil, err
	}

	// http.Client handles redirects at a layer above the http.RoundTripper
	// abstraction, so to avoid forwarding Authorization headers to places
	// we are redirected, only set it when the authorization header matches
	// the registry with which we are interacting.
	if in.Host == bt.registry.RegistryStr() {
		in.Header.Set("Authorization", hdr)
	}
	in.Header.Set("User-Agent", transportName)

	// TODO(mattmoor): On 401s perform a single refresh() and retry.
	return bt.inner.RoundTrip(in)
}

func (bt *bearerTransport) refresh() error {
	u, err := url.Parse(bt.realm)
	if err != nil {
		return err
	}
	b := &basicTransport{
		inner:  bt.inner,
		auth:   bt.basic,
		target: u.Host,
	}
	client := http.Client{Transport: b}

	u.RawQuery = url.Values{
		"scope":   []string{bt.scope},
		"service": []string{bt.service},
	}.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the response into a Bearer authenticator
	bearer := &authn.Bearer{}
	if err := json.Unmarshal(content, bearer); err != nil {
		return err
	}
	// Replace our old bearer authenticator (if we had one) with our newly refreshed authenticator.
	bt.bearer = bearer
	return nil
}

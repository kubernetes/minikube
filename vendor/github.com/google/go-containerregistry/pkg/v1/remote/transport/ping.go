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
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

type challenge string

const (
	anonymous challenge = "anonymous"
	basic     challenge = "basic"
	bearer    challenge = "bearer"
)

type pingResp struct {
	challenge challenge

	// Following the challenge there are often key/value pairs
	// e.g. Bearer service="gcr.io",realm="https://auth.gcr.io/v36/tokenz"
	parameters map[string]string
}

func (c challenge) Canonical() challenge {
	return challenge(strings.ToLower(string(c)))
}

func parseChallenge(suffix string) map[string]string {
	kv := make(map[string]string)
	for _, token := range strings.Split(suffix, ",") {
		// Trim any whitespace around each token.
		token = strings.Trim(token, " ")

		// Break the token into a key/value pair
		if parts := strings.SplitN(token, "=", 2); len(parts) == 2 {
			// Unquote the value, if it is quoted.
			kv[parts[0]] = strings.Trim(parts[1], `"`)
		} else {
			// If there was only one part, treat is as a key with an empty value
			kv[token] = ""
		}
	}
	return kv
}

func ping(reg name.Registry, t http.RoundTripper) (*pingResp, error) {
	client := http.Client{Transport: t}

	url := fmt.Sprintf("%s://%s/v2/", reg.Scheme(), reg.Name())
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// If we get a 200, then no authentication is needed.
		return &pingResp{challenge: anonymous}, nil
	case http.StatusUnauthorized:
		wac := resp.Header.Get(http.CanonicalHeaderKey("WWW-Authenticate"))
		if parts := strings.SplitN(wac, " ", 2); len(parts) == 2 {
			// If there are two parts, then parse the challenge parameters.
			return &pingResp{
				challenge:  challenge(parts[0]).Canonical(),
				parameters: parseChallenge(parts[1]),
			}, nil
		}
		// Otherwise, just return the challenge without parameters.
		return &pingResp{
			challenge: challenge(wac).Canonical(),
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized HTTP status: %v", resp.Status)
	}
}

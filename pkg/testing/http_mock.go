/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package testing

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	OctetStream = "application/octet-stream"
	JSON        = "application/json; charset=utf-8"
	TEXT        = "text/plain"
)

var DefaultRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func ResetDefaultRoundTripper() {
	http.DefaultClient.Transport = DefaultRoundTripper
}

// MockRoundTripper mocks HTTP requests and allows to return canned responses
type MockRoundTripper struct {
	delegate        http.RoundTripper
	responses       map[string]*CannedResponse
	verbose         bool
	allowDelegation bool
}

type ResponseType int

const (
	ServeString ResponseType = iota
	ServeFile
)

type CannedResponse struct {
	ResponseType ResponseType
	Response     string
	ContentType  string
}

func NewMockRoundTripper() *MockRoundTripper {
	return &MockRoundTripper{
		delegate:        DefaultRoundTripper,
		responses:       make(map[string]*CannedResponse),
		verbose:         false,
		allowDelegation: false,
	}
}

func (t *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.verbose {
		fmt.Println(fmt.Sprintf("MockRoundTripper received HTTP request '%s'", req.URL.String()))
	}

	for url, cannedResponse := range t.responses {
		matched, _ := regexp.Match(fmt.Sprintf("^%s$", url), []byte(req.URL.String()))
		if !matched {
			if t.verbose {
				fmt.Println(fmt.Sprintf("Not get registered response for '%s'", []byte(req.URL.String())))
			}
			continue
		}

		response := t.createResponseFor(req, cannedResponse.ContentType)

		switch cannedResponse.ResponseType {
		case ServeString:
			response.Body = ioutil.NopCloser(strings.NewReader(cannedResponse.Response))
		case ServeFile:
			file, err := os.Open(cannedResponse.Response)
			if err != nil {
				panic(err)
			}
			response.Body = ioutil.NopCloser(bufio.NewReader(file))
		default:
			panic("Unknown canned response type")

		}
		if t.verbose {
			fmt.Println(fmt.Sprintf("Returning canned response for HTTP request '%s'", req.URL.String()))
		}
		return response, nil
	}

	// Otherwise delegate
	if t.verbose {
		fmt.Println(fmt.Sprintf("Delegating '%s'", req.URL.String()))
	}

	if !t.allowDelegation {
		panic(fmt.Sprintf("Not allowed to delegate '%s'", req.URL.String()))
	}

	return t.delegate.RoundTrip(req)
}

func (t *MockRoundTripper) RegisterResponse(url string, response *CannedResponse) {
	t.responses[url] = response
}

func (t *MockRoundTripper) Verbose(verbose bool) {
	t.verbose = verbose
}

func (t *MockRoundTripper) AllowDelegation(delegate bool) {
	t.allowDelegation = delegate
}

func (t *MockRoundTripper) createResponseFor(req *http.Request, contentType string) *http.Response {
	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: http.StatusOK,
	}
	response.Header.Set("Content-Type", contentType)
	return response
}

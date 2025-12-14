// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

type TransportGetter func(t *testing.T, config *config.Config) *http.Transport

func HTTPClientConfigurationTest_basic(t *testing.T, transport *http.Transport) {
	t.Helper()

	if a, e := transport.MaxIdleConns, awshttp.DefaultHTTPTransportMaxIdleConns; a != e {
		t.Errorf("expected MaxIdleConns to be %d, got %d", e, a)
	}
	if a, e := transport.MaxIdleConnsPerHost, awshttp.DefaultHTTPTransportMaxIdleConnsPerHost; a != e {
		t.Errorf("expected MaxIdleConnsPerHost to be %d, got %d", e, a)
	}
	if a, e := transport.IdleConnTimeout, awshttp.DefaultHTTPTransportIdleConnTimeout; a != e {
		t.Errorf("expected IdleConnTimeout to be %s, got %s", e, a)
	}
	if a, e := transport.TLSHandshakeTimeout, awshttp.DefaultHTTPTransportTLSHandleshakeTimeout; a != e {
		t.Errorf("expected TLSHandshakeTimeout to be %s, got %s", e, a)
	}
	if a, e := transport.ExpectContinueTimeout, awshttp.DefaultHTTPTransportExpectContinueTimeout; a != e {
		t.Errorf("expected ExpectContinueTimeout to be %s, got %s", e, a)
	}
	if !transport.ForceAttemptHTTP2 {
		t.Error("expected ForceAttemptHTTP2 to be true, got false")
	}
	if transport.DisableKeepAlives {
		t.Error("expected DisableKeepAlives to be false, got true")
	}

	tlsConfig := transport.TLSClientConfig
	if a, e := int(tlsConfig.MinVersion), tls.VersionTLS12; a != e {
		t.Errorf("expected tlsConfig.MinVersion to be %d, got %d", e, a)
	}
	if tlsConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be false, got true")
	}
}

func HTTPClientConfigurationTest_insecureHTTPS(t *testing.T, transport *http.Transport) {
	t.Helper()

	tlsConfig := transport.TLSClientConfig
	if !tlsConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true, got false")
	}
}

type proxyCase struct {
	url           string
	expectedProxy string
}

func HTTPClientConfigurationTest_proxy(t *testing.T, getter TransportGetter) {
	t.Helper()

	// Go supports both the upper- and lower-case versions of the proxy environment variables
	testcases := map[string]struct {
		config               config.Config
		environmentVariables map[string]string
		urls                 []proxyCase
	}{
		"no config": {
			config: config.Config{},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config empty string": {
			config: config.Config{
				HTTPProxy: aws.String(""),
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config Legacy": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeLegacy,
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config Separate": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeSeparate,
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPSProxy config": {
			config: config.Config{
				HTTPSProxy: aws.String("http://https-proxy.test:1234"),
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config HTTPSProxy config": {
			config: config.Config{
				HTTPProxy:  aws.String("http://http-proxy.test:1234"),
				HTTPSProxy: aws.String("http://https-proxy.test:1234"),
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config HTTPSProxy config empty string Legacy": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPSProxy:    aws.String(""),
				HTTPProxyMode: config.HTTPProxyModeLegacy,
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPSProxy config HTTPProxy config empty string": {
			config: config.Config{
				HTTPProxy:  aws.String(""),
				HTTPSProxy: aws.String("http://https-proxy.test:1234"),
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config HTTPSProxy config NoProxy config": {
			config: config.Config{
				HTTPProxy:  aws.String("http://http-proxy.test:1234"),
				HTTPSProxy: aws.String("http://https-proxy.test:1234"),
				NoProxy:    "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTP_PROXY envvar": {
			config: config.Config{},
			environmentVariables: map[string]string{
				"HTTP_PROXY": "http://http-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"http_proxy envvar": {
			config: config.Config{},
			environmentVariables: map[string]string{
				"http_proxy": "http://http-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPS_PROXY envvar": {
			config: config.Config{},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"https_proxy envvar": {
			config: config.Config{},
			environmentVariables: map[string]string{
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config HTTPS_PROXY envvar": {
			config: config.Config{
				HTTPProxy: aws.String("http://http-proxy.test:1234"),
			},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config https_proxy envvar": {
			config: config.Config{
				HTTPProxy: aws.String("http://http-proxy.test:1234"),
			},
			environmentVariables: map[string]string{
				"https_proxy": "http://https-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config NO_PROXY envvar Legacy": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeLegacy,
			},
			environmentVariables: map[string]string{
				"NO_PROXY": "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config NO_PROXY envvar Separate": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeSeparate,
			},
			environmentVariables: map[string]string{
				"NO_PROXY": "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config no_proxy envvar Legacy": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeLegacy,
			},
			environmentVariables: map[string]string{
				"no_proxy": "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config no_proxy envvar Separate": {
			config: config.Config{
				HTTPProxy:     aws.String("http://http-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeSeparate,
			},
			environmentVariables: map[string]string{
				"no_proxy": "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTP_PROXY envvar HTTPS_PROXY envvar NO_PROXY envvar": {
			config: config.Config{},
			environmentVariables: map[string]string{
				"HTTP_PROXY":  "http://http-proxy.test:1234",
				"HTTPS_PROXY": "http://https-proxy.test:1234",
				"NO_PROXY":    "dont-proxy.test",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://http-proxy.test:1234",
				},
				{
					url:           "http://dont-proxy.test",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://https-proxy.test:1234",
				},
				{
					url:           "https://dont-proxy.test",
					expectedProxy: "",
				},
			},
		},

		"HTTPProxy config overrides HTTP_PROXY envvar Legacy": {
			config: config.Config{
				HTTPProxy:     aws.String("http://config-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeLegacy,
			},
			environmentVariables: map[string]string{
				"HTTP_PROXY": "http://envvar-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
			},
		},

		"HTTPProxy config overrides HTTP_PROXY envvar Separate": {
			config: config.Config{
				HTTPProxy:     aws.String("http://config-proxy.test:1234"),
				HTTPProxyMode: config.HTTPProxyModeSeparate,
			},
			environmentVariables: map[string]string{
				"HTTP_PROXY": "http://envvar-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
				{
					url:           "https://example.com",
					expectedProxy: "",
				},
			},
		},

		"HTTPSProxy config overrides HTTPS_PROXY envvar": {
			config: config.Config{
				HTTPSProxy: aws.String("http://config-proxy.test:1234"),
			},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://envvar-proxy.test:1234",
			},
			urls: []proxyCase{
				{
					url:           "http://example.com",
					expectedProxy: "",
				},
				{
					url:           "https://example.com",
					expectedProxy: "http://config-proxy.test:1234",
				},
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testcase.environmentVariables {
				t.Setenv(k, v)
			}

			transport := getter(t, &testcase.config)
			proxy := transport.Proxy

			for _, url := range testcase.urls {
				req, _ := http.NewRequest("GET", url.url, nil)
				pUrl, err := proxy(req)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if url.expectedProxy != "" {
					if pUrl == nil {
						t.Errorf("expected proxy for %q, got none", url.url)
					} else if pUrl.String() != url.expectedProxy {
						t.Errorf("expected proxy %q for %q, got %q", url.expectedProxy, url.url, pUrl.String())
					}
				} else {
					if pUrl != nil {
						t.Errorf("expected no proxy for %q, got %q", url.url, pUrl.String())
					}
				}
			}
		})
	}
}

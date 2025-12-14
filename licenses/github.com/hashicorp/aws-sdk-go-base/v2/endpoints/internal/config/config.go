// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/expand"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"golang.org/x/net/http/httpproxy"
)

type ProxyMode int

const (
	HTTPProxyModeLegacy ProxyMode = iota
	HTTPProxyModeSeparate
)

type Config struct {
	AccessKey                      string
	AllowedAccountIds              []string
	APNInfo                        *APNInfo
	AssumeRole                     []AssumeRole
	AssumeRoleWithWebIdentity      *AssumeRoleWithWebIdentity
	Backoff                        retry.BackoffDelayer
	CallerDocumentationURL         string
	CallerName                     string
	CustomCABundle                 string
	EC2MetadataServiceEnableState  imds.ClientEnableState
	EC2MetadataServiceEndpoint     string
	EC2MetadataServiceEndpointMode string
	ForbiddenAccountIds            []string
	HTTPClient                     *http.Client
	HTTPProxy                      *string
	HTTPSProxy                     *string
	IamEndpoint                    string
	Insecure                       bool
	Logger                         logging.Logger
	MaxBackoff                     time.Duration
	MaxRetries                     int
	NoProxy                        string
	Profile                        string
	HTTPProxyMode                  ProxyMode
	Region                         string
	RetryMode                      aws.RetryMode
	SecretKey                      string
	SharedCredentialsFiles         []string
	SharedConfigFiles              []string
	SkipCredsValidation            bool
	SkipRequestingAccountId        bool
	SsoEndpoint                    string
	StsEndpoint                    string
	StsRegion                      string
	SuppressDebugLog               bool
	Token                          string
	TokenBucketRateLimiterCapacity int
	UseDualStackEndpoint           bool
	UseFIPSEndpoint                bool
	UserAgent                      UserAgentProducts
}

type AssumeRole struct {
	RoleARN           string
	Duration          time.Duration
	ExternalID        string
	Policy            string
	PolicyARNs        []string
	SessionName       string
	SourceIdentity    string
	Tags              map[string]string
	TransitiveTagKeys []string
}

func (c Config) CustomCABundleReader() (*bytes.Reader, error) {
	if c.CustomCABundle == "" {
		return nil, nil
	}
	bundleFile, err := expand.FilePath(c.CustomCABundle)
	if err != nil {
		return nil, fmt.Errorf("expanding custom CA bundle: %w", err)
	}
	bundle, err := os.ReadFile(bundleFile)
	if err != nil {
		return nil, fmt.Errorf("reading custom CA bundle: %w", err)
	}
	return bytes.NewReader(bundle), nil
}

// HTTPTransportOptions returns functional options that configures an http.Transport.
// The returned options function is called on both AWS SDKv1 and v2 default HTTP clients.
func (c Config) HTTPTransportOptions() (func(*http.Transport), error) {
	var err error
	var httpProxyUrl *url.URL
	if c.HTTPProxy != nil {
		httpProxyUrl, err = url.Parse(aws.ToString(c.HTTPProxy))
		if err != nil {
			return nil, fmt.Errorf("parsing HTTP proxy URL: %w", err)
		}
	}
	var httpsProxyUrl *url.URL
	if c.HTTPSProxy != nil {
		httpsProxyUrl, err = url.Parse(aws.ToString(c.HTTPSProxy))
		if err != nil {
			return nil, fmt.Errorf("parsing HTTPS proxy URL: %w", err)
		}
	}

	opts := func(tr *http.Transport) {
		tr.MaxIdleConnsPerHost = awshttp.DefaultHTTPTransportMaxIdleConnsPerHost

		tlsConfig := tr.TLSClientConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
			tr.TLSClientConfig = tlsConfig
		}

		if c.Insecure {
			tr.TLSClientConfig.InsecureSkipVerify = true
		}

		proxyConfig := httpproxy.FromEnvironment()
		if httpProxyUrl != nil {
			proxyConfig.HTTPProxy = httpProxyUrl.String()
			if c.HTTPProxyMode == HTTPProxyModeLegacy && proxyConfig.HTTPSProxy == "" {
				proxyConfig.HTTPSProxy = httpProxyUrl.String()
			}
		}
		if httpsProxyUrl != nil {
			proxyConfig.HTTPSProxy = httpsProxyUrl.String()
		}
		if c.NoProxy != "" {
			proxyConfig.NoProxy = c.NoProxy
		}
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxyConfig.ProxyFunc()(req.URL)
		}
	}

	return opts, nil
}

func (c Config) ValidateProxySettings(diags *diag.Diagnostics) {
	if c.HTTPProxy != nil {
		if _, err := url.Parse(aws.ToString(c.HTTPProxy)); err != nil {
			*diags = diags.AddError(
				"Invalid HTTP Proxy",
				fmt.Sprintf("Unable to parse URL: %s", err),
			)
		}
	}

	if c.HTTPSProxy != nil {
		if _, err := url.Parse(aws.ToString(c.HTTPSProxy)); err != nil {
			*diags = diags.AddError(
				"Invalid HTTPS Proxy",
				fmt.Sprintf("Unable to parse URL: %s", err),
			)
		}
	}

	if c.HTTPProxy != nil && *c.HTTPProxy != "" && c.HTTPSProxy == nil && os.Getenv("HTTPS_PROXY") == "" && os.Getenv("https_proxy") == "" {
		if c.HTTPProxyMode == HTTPProxyModeLegacy {
			*diags = diags.Append(
				missingHttpsProxyLegacyWarningDiag(aws.ToString(c.HTTPProxy)),
			)
		} else {
			*diags = diags.Append(
				missingHttpsProxyWarningDiag(),
			)
		}
	}
}

const (
	missingHttpsProxyWarningSummary   = "Missing HTTPS Proxy"
	missingHttpsProxyDetailProblem    = "An HTTP proxy was set but no HTTPS proxy was."
	missingHttpsProxyDetailResolution = "To specify no proxy for HTTPS, set the HTTPS to an empty string."
)

func missingHttpsProxyLegacyWarningDiag(s string) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		missingHttpsProxyWarningSummary,
		fmt.Sprintf(
			missingHttpsProxyDetailProblem+" Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
				missingHttpsProxyDetailResolution,
			s,
		),
	)
}

func missingHttpsProxyWarningDiag() diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		missingHttpsProxyWarningSummary,
		missingHttpsProxyDetailProblem+"\n\n"+
			missingHttpsProxyDetailResolution,
	)
}

func (c Config) ResolveSharedConfigFiles() ([]string, error) {
	v, err := expand.FilePaths(c.SharedConfigFiles)
	if err != nil {
		return []string{}, fmt.Errorf("expanding shared config files: %w", err)
	}
	return v, nil
}

func (c Config) ResolveSharedCredentialsFiles() ([]string, error) {
	v, err := expand.FilePaths(c.SharedCredentialsFiles)
	if err != nil {
		return []string{}, fmt.Errorf("expanding shared credentials files: %w", err)
	}
	return v, nil
}

// VerifyAccountIDAllowed verifies an account ID is not explicitly forbidden
// or omitted from an allow list, if configured.
//
// If the AllowedAccountIds and ForbiddenAccountIds fields are both empty, this
// function will return nil.
func (c Config) VerifyAccountIDAllowed(accountID string) error {
	if len(c.ForbiddenAccountIds) > 0 {
		if slices.Contains(c.ForbiddenAccountIds, accountID) {
			return fmt.Errorf("AWS account ID not allowed: %s", accountID)
		}
	}
	if len(c.AllowedAccountIds) > 0 {
		found := slices.Contains(c.AllowedAccountIds, accountID)
		if !found {
			return fmt.Errorf("AWS account ID not allowed: %s", accountID)
		}
	}
	return nil
}

type AssumeRoleWithWebIdentity struct {
	RoleARN              string
	Duration             time.Duration
	Policy               string
	PolicyARNs           []string
	SessionName          string
	WebIdentityToken     string
	WebIdentityTokenFile string
}

func (c AssumeRoleWithWebIdentity) resolveWebIdentityTokenFile() (string, error) {
	v, err := expand.FilePath(c.WebIdentityTokenFile)
	if err != nil {
		return "", fmt.Errorf("expanding web identity token file: %w", err)
	}
	return v, nil
}

func (c AssumeRoleWithWebIdentity) HasValidTokenSource() bool {
	return c.WebIdentityToken != "" || c.WebIdentityTokenFile != ""
}

// Implements `stscreds.IdentityTokenRetriever`
func (c AssumeRoleWithWebIdentity) GetIdentityToken() ([]byte, error) {
	if c.WebIdentityToken != "" {
		return []byte(c.WebIdentityToken), nil
	}
	webIdentityTokenFile, err := c.resolveWebIdentityTokenFile()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(webIdentityTokenFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read file at %s: %w", webIdentityTokenFile, err)
	}

	return b, nil
}

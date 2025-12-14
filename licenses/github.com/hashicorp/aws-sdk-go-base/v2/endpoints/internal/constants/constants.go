// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package constants

const (
	// AppendUserAgentEnvVar is a conventionally used environment variable
	// containing additional HTTP User-Agent information.
	// If present and its value is non-empty, it is directly appended to the
	// User-Agent header for HTTP requests.
	AppendUserAgentEnvVar = "TF_APPEND_USER_AGENT"

	// Maximum network retries.
	// We depend on the AWS Go SDK DefaultRetryer exponential backoff.
	// Ensure that if the AWS Config MaxRetries is set high (which it is by
	// default), that we only retry for a few seconds with typically
	// unrecoverable network errors, such as DNS lookup failures.
	MaxNetworkRetryCount = 9
)

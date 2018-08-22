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

package remote

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Error implements error to support the following error specification:
// https://github.com/docker/distribution/blob/master/docs/spec/api.md#errors
type Error struct {
	Errors []Diagnostic `json:"errors,omitempty"`
}

// Check that Error implements error
var _ error = (*Error)(nil)

// Error implements error
func (e *Error) Error() string {
	switch len(e.Errors) {
	case 0:
		return "<empty remote.Error response>"
	case 1:
		return e.Errors[0].String()
	default:
		var errors []string
		for _, d := range e.Errors {
			errors = append(errors, d.String())
		}
		return fmt.Sprintf("multiple errors returned: %s",
			strings.Join(errors, ";"))
	}
}

// Diagnostic represents a single error returned by a Docker registry interaction.
type Diagnostic struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
}

// String stringifies the Diagnostic
func (d Diagnostic) String() string {
	return fmt.Sprintf("%s: %q", d.Code, d.Message)
}

// ErrorCode is an enumeration of supported error codes.
type ErrorCode string

// The set of error conditions a registry may return:
// https://github.com/docker/distribution/blob/master/docs/spec/api.md#errors-2
const (
	BlobUnknownErrorCode         ErrorCode = "BLOB_UNKNOWN"
	BlobUploadInvalidErrorCode   ErrorCode = "BLOB_UPLOAD_INVALID"
	BlobUploadUnknownErrorCode   ErrorCode = "BLOB_UPLOAD_UNKNOWN"
	DigestInvalidErrorCode       ErrorCode = "DIGEST_INVALID"
	ManifestBlobUnknownErrorCode ErrorCode = "MANIFEST_BLOB_UNKNOWN"
	ManifestInvalidErrorCode     ErrorCode = "MANIFEST_INVALID"
	ManifestUnknownErrorCode     ErrorCode = "MANIFEST_UNKNOWN"
	ManifestUnverifiedErrorCode  ErrorCode = "MANIFEST_UNVERIFIED"
	NameInvalidErrorCode         ErrorCode = "NAME_INVALID"
	NameUnknownErrorCode         ErrorCode = "NAME_UNKNOWN"
	SizeInvalidErrorCode         ErrorCode = "SIZE_INVALID"
	TagInvalidErrorCode          ErrorCode = "TAG_INVALID"
	UnauthorizedErrorCode        ErrorCode = "UNAUTHORIZED"
	DeniedErrorCode              ErrorCode = "DENIED"
	UnsupportedErrorCode         ErrorCode = "UNSUPPORTED"
)

func CheckError(resp *http.Response, codes ...int) error {
	for _, code := range codes {
		if resp.StatusCode == code {
			// This is one of the supported status codes.
			return nil
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// https://github.com/docker/distribution/blob/master/docs/spec/api.md#errors
	var structuredError Error
	if err := json.Unmarshal(b, &structuredError); err != nil {
		// If the response isn't an unstructured error, then return some
		// reasonable error response containing the response body.
		return fmt.Errorf("unsupported status code %d; body: %s", resp.StatusCode, string(b))
	}
	return &structuredError
}

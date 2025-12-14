// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfawserr

import (
	"slices"
	"strings"

	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/errs"
)

// ErrCodeEquals returns true if the error matches all these conditions:
//   - err is of type smithy.APIError
//   - APIError.ErrorCode() equals one of the passed codes
func ErrCodeEquals(err error, codes ...string) bool {
	if apiErr, ok := errs.As[smithy.APIError](err); ok {
		if slices.Contains(codes, apiErr.ErrorCode()) {
			return true
		}
	}
	return false
}

// ErrCodeContains returns true if the error matches all these conditions:
//   - err is of type smithy.APIError
//   - APIError.ErrorCode() contains code
func ErrCodeContains(err error, code string) bool {
	if apiErr, ok := errs.As[smithy.APIError](err); ok {
		return strings.Contains(apiErr.ErrorCode(), code)
	}
	return false
}

// ErrMessageContains returns true if the error matches all these conditions:
//   - err is of type smithy.APIError
//   - APIError.ErrorCode() equals code
//   - APIError.ErrorMessage() contains message
func ErrMessageContains(err error, code string, message string) bool {
	if apiErr, ok := errs.As[smithy.APIError](err); ok {
		return apiErr.ErrorCode() == code && strings.Contains(apiErr.ErrorMessage(), message)
	}
	return false
}

// ErrHTTPStatusCodeEquals returns true if the error matches all these conditions:
//   - err is of type smithyhttp.ResponseError
//   - ResponseError.HTTPStatusCode() equals one of the passed status codes
func ErrHTTPStatusCodeEquals(err error, statusCodes ...int) bool {
	if respErr, ok := errs.As[*smithyhttp.ResponseError](err); ok {
		if slices.Contains(statusCodes, respErr.HTTPStatusCode()) {
			return true
		}
	}
	return false
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// +build !windows

package url

import (
	"net/url"
)

func parse(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

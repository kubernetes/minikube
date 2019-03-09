// Copyright 2018, Rapid7, Inc.
// License: BSD-3-clause
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// * Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
// * Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software
// without specific prior written permission.
package proxy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const maxUint16 = (1 << 16) - 1

/*
Parse the optionally valid URL string. Should the URL not contain a Scheme, an empty one will be provided.
Should all hope be lost after that, expect error to be populated.
Params:
	rawUrl: An optionally valid URL
Returns:
	url.URL: The parsed URL if we managed to construct a valid one.
	error: If URL was invalid, even after providing a Scheme.
*/
func ParseURL(rawUrl string, defaultScheme string) (*url.URL, error) {
	rawUrl = prefixScheme(strings.TrimSpace(rawUrl), defaultScheme)
	return url.Parse(rawUrl)
}

/*
Sanitize the given target URL string to include only the Scheme, Host, and Port.
Params:
	targetUrl: An optionally valid URL
Returns:
	If any of the following exists in the given URL string, it will be omitted from the return value:
		* Username
		* Password
		* Query params
		* Fragment
	The given URL string need not be a valid URL
*/
func ParseTargetURL(targetUrl, defaultScheme string) *url.URL {
	parsedUrl, err := ParseURL(targetUrl, defaultScheme)
	if err != nil {
		return &url.URL{Host: targetUrlWildcard}
	}
	if parsedUrl.Host == "" {
		parsedUrl.Host = targetUrlWildcard
	}
	parsedUrl.User = nil
	parsedUrl.Fragment = ""
	parsedUrl.RawQuery = ""
	parsedUrl.ForceQuery = false
	parsedUrl.Opaque = ""
	parsedUrl.Path = ""
	parsedUrl.RawPath = ""
	return parsedUrl
}

/*
Split the optionally valid URL into a host and port.
Should the URL be invalid or have no Host entry, "", 0, err will be returned.
Params:
	u: An optionally valid URL
Returns:
	Returns host, port, err.
		If URL is valid: host, port, nil
		If URL is invalid: "", 0, error
*/
func SplitHostPort(u *url.URL) (host string, port uint16, err error) {
	if u == nil {
		return "", 0, newSplitHostPortError("nil", errors.New("nil URL"))
	}
	host = strings.TrimSpace(u.Host)
	portStr := ""
	// Find last colon.
	i := strings.LastIndex(host, ":")
	if i >= 0 {
		// Special case for IPv6, we may be within brackets
		if len(host) > i && strings.Index(host[i:], "]") < 0 {
			portStr = host[i+1:]
			host = host[:i]
		}
	}
	if portStr != "" {
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return "", 0, newSplitHostPortError(u.Host, err)
		} else if port64 > maxUint16 {
			return "", 0, newSplitHostPortError(u.Host, errors.New(fmt.Sprintf("%d > %d", port64, maxUint16)))
		}
		port = uint16(port64)
	}
	return host, port, nil
}

/*
Determines if the given host string (either hostname or IP) references a loop back address.
The following are considered loop back addresses:
	* 127.0.0.1/8
	* [::1]
	* localhost
Params:
	host: Hostname or IP.
Returns:
	Returns true if the host references a loop back address, false otherwise.
*/
//noinspection SpellCheckingInspection
func IsLoopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if host == "localhost" {
		return true
	}
	// Get raw IPv6 value if present in the host
	if len(host) >= 2 && host[0] == '[' && host[len(host)-1] == ']' {
		host = host[1 : len(host)-1]
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

/*
Prefix the given rawURL string with a scheme.
If a defaultScheme is provided, and no scheme exists for rawURL, it will be used.
For example:
	("test:8080", "https") -> "https://test:8080"
	("https://8080", "gopher") -> "https://test:8080"
	("test:8080", "") -> "//test:8080"
Params:
	rawURL: The raw URL as a string which may or may not have a scheme.
	defaultScheme: Optional. The scheme to inject if no scheme is present in rawURL.
Returns:
	Returns the given rawURL with a scheme.
*/
func prefixScheme(rawURL string, defaultScheme string) string {
	if strings.HasPrefix(rawURL, "//") {
		// Empty scheme, but a default one was provided
		if defaultScheme != "" {
			rawURL = defaultScheme + ":" + rawURL
		}
	} else if strings.Index(rawURL, "://") < 0 {
		// No scheme specified at all, place an empty one
		rawURL = "//" + rawURL
		// A default scheme was provided, prefix it
		if defaultScheme != "" {
			rawURL = defaultScheme + ":" + rawURL
		}
	}
	return rawURL
}

func newSplitHostPortError(host string, err error) *url.Error {
	return &url.Error{Op: "SplitHostPort", URL: host, Err: err}
}

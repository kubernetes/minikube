// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import "net/url"

// RedactURL is a port of url.Redacted from the standard library,
// which is like url.String but replaces any password with "redacted".
// Only the password in u.URL is redacted. This allows the library
// to maintain compatibility with go1.14.
// This port was also extended to redact all "sshkey" from URL query parameter
// and replace them with "redacted".
func RedactURL(u *url.URL) string {
	if u == nil {
		return ""
	}

	ru := *u
	if _, has := ru.User.Password(); has {
		ru.User = url.UserPassword(ru.User.Username(), "redacted")
	}
	q := ru.Query()
	if q.Has("sshkey") {
		values := q["sshkey"]
		for i := range values {
			values[i] = "redacted"
		}
		ru.RawQuery = q.Encode()
	}
	return ru.String()
}

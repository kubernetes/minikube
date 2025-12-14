// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"regexp"
	"unsafe"
)

// IAM Unique ID prefixes from
// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-unique-ids
var UniqueIDRegex = regexp.MustCompile(`(A3T[A-Z0-9]` +
	`|ABIA` + // STS service bearer token
	`|ACCA` + // Context-specific credential
	`|AGPA` + // User group
	`|AIDA` + // IAM user
	`|AIPA` + // EC2 instance profile
	`|AKIA` + // Access key
	`|ANPA` + // Managed policy
	`|ANVA` + // Version in a managed policy
	`|APKA` + // Public key
	`|AROA` + // Role
	`|ASCA` + // Certificate
	`|ASIA` + // STS temporary access key
	`)[A-Z0-9]{16,}`)

const (
	unmaskedFirst = 4
	unmaskedLast  = 4
)

func MaskAWSAccessKey(field []byte) []byte {
	field = UniqueIDRegex.ReplaceAllFunc(field, func(s []byte) []byte {
		return partialMaskString(s, unmaskedFirst, unmaskedLast)
	})
	return field
}

func MaskAWSSensitiveValues(field string) string {
	b := unsafe.Slice(unsafe.StringData(field), len(field))
	b = MaskAWSAccessKey(b)
	MaskAWSSecretKeys(b)
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// MaskAWSSecretKeys masks likely AWS secret access keys in the input.
// See https://aws.amazon.com/blogs/security/a-safer-way-to-distribute-aws-credentials-to-ec2/:
// "Find me 40-character, base-64 strings that don’t have any base 64 characters immediately before or after".
func MaskAWSSecretKeys(in []byte) {
	const (
		secretKeyLen = 40
	)
	len := len(in)
	base64Characters := 0

	for i := range len {
		b := in[i]

		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '/' || b == '+' || b == '=' {
			// base64 character.
			base64Characters++
		} else {
			if base64Characters == secretKeyLen {
				for j := (i - secretKeyLen) + unmaskedFirst; j < i-unmaskedLast; j++ {
					in[j] = '*'
				}
			}

			base64Characters = 0
		}
	}

	if base64Characters == secretKeyLen {
		for j := (len - secretKeyLen) + unmaskedFirst; j < len-unmaskedLast; j++ {
			in[j] = '*'
		}
	}
}

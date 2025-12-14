// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

func partialMaskString(s []byte, first, last int) []byte {
	l := len(s)
	result := make([]byte, 0, l)
	result = append(result, s[0:first]...)
	for range l - first - last {
		result = append(result, '*')
	}
	result = append(result, s[l-last:]...)
	return result
}

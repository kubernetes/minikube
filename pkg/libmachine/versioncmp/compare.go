// Package versioncmp provides functions for comparing version strings.
//
// Version strings are dot-separated integers with an optional
// pre-release suffix. A pre-release suffix is an arbitrary string with a
// leading dash character. All functions ignore these suffixes, so "1.2" and
// "1.2-rc" are considered equivalent.
package versioncmp

import (
	"strconv"
	"strings"
)

const (
	rcString  = "-rc"
	ceEdition = "-ce"
)

// compare compares two versions of Docker to decipher which came first.
//
// compare returns -1 if v1 < v2, 1 if v1 > v2, 0 otherwise.
func compare(v1, v2 string) int {
	// Replace RC string with "." to make the RC number appear as simply
	// another sub-version.
	v1 = strings.Replace(v1, rcString, ".", -1)
	v2 = strings.Replace(v2, rcString, ".", -1)

	// All releases before the community edition (differentiated by
	// presence of the "ce" string in the version string) are "less than"
	// any community edition release (first occurring in March 2017).
	if strings.Contains(v1, ceEdition) && !strings.Contains(v2, ceEdition) {
		return 1
	}
	if !strings.Contains(v1, ceEdition) && strings.Contains(v2, ceEdition) {
		return -1
	}

	// Without this tag, both are pre-CE versions.
	if !strings.Contains(v1, ceEdition) && !strings.Contains(v2, ceEdition) {
		return compareNumeric(v1, v2)
	}

	return compareCE(v1, v2)
}

// compareCE ("Community Edition") will differentiate between versions of
// Docker that use the versioning scheme
// {{release-year}}.{{release-month}}-{{ce|ee}}-{{rcnum|""}}
//
// This will be every release after 1.13.1.
func compareCE(v1, v2 string) int {
	return compareNumeric(
		strings.Replace(v1, ceEdition, "", -1),
		strings.Replace(v2, ceEdition, "", -1),
	)
}

// compareNumeric compares two version that use pre-17.03 Docker.
//
// Non-numeric segments in either argument are considered equal, so
// compare("1.a", "1.b") == 0, but compare("2.a", "1.b") == 1.
func compareNumeric(v1, v2 string) int {
	if n := strings.IndexByte(v1, '-'); n != -1 {
		v1 = v1[:n]
	}
	if n := strings.IndexByte(v2, '-'); n != -1 {
		v2 = v2[:n]
	}
	var (
		currTab  = strings.Split(v1, ".")
		otherTab = strings.Split(v2, ".")
	)

	max := len(currTab)
	if len(otherTab) > max {
		max = len(otherTab)
	}
	for i := 0; i < max; i++ {
		var currInt, otherInt int

		if len(currTab) > i {
			currInt, _ = strconv.Atoi(currTab[i])
		}
		if len(otherTab) > i {
			otherInt, _ = strconv.Atoi(otherTab[i])
		}
		if currInt > otherInt {
			return 1
		}
		if otherInt > currInt {
			return -1
		}
	}
	return 0
}

// LessThan checks if a version is less than another.
func LessThan(v, other string) bool {
	return compare(v, other) == -1
}

// LessThanOrEqualTo checks if a version is less than or equal to another.
func LessThanOrEqualTo(v, other string) bool {
	return compare(v, other) <= 0
}

// GreaterThan checks if a version is greater than another.
func GreaterThan(v, other string) bool {
	return compare(v, other) == 1
}

// GreaterThanOrEqualTo checks if a version is greater than or equal to
// another.
func GreaterThanOrEqualTo(v, other string) bool {
	return compare(v, other) >= 0
}

// Equal checks if a version is equal to another.
func Equal(v, other string) bool {
	return compare(v, other) == 0
}

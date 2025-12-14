// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import "slices"

// ApplyToAll returns a new slice containing the results of applying the function `f` to each element of the original slice `s`.
func ApplyToAll[T, U any](s []T, f func(T) U) []U {
	v := make([]U, len(s))

	for i, e := range s {
		v[i] = f(e)
	}

	return v
}

type FilterFunc[T any] func(T) bool

// Filter returns a new slice containing all values that return `true` for the filter function `f`
func Filter[T any](s []T, f FilterFunc[T]) []T {
	v := make([]T, 0, len(s))

	for _, e := range s {
		if f(e) {
			v = append(v, e)
		}
	}

	return slices.Clip(v)
}

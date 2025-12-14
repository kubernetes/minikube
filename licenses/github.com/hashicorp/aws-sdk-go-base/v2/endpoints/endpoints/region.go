// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package endpoints

// Region represents an AWS Region.
// See https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/regions.html.
type Region struct {
	id          string
	description string
}

// ID returns the Region's identifier.
func (r Region) ID() string {
	return r.id
}

// Description returns the Region's description.
func (r Region) Description() string {
	return r.description
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package endpoints

import (
	"maps"
	"regexp"
)

// Partition represents an AWS partition.
// See https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html.
type Partition struct {
	id          string
	name        string
	dnsSuffix   string
	regionRegex *regexp.Regexp
	regions     map[string]Region
	services    map[string]Service
}

// ID returns the identifier of the partition.
func (p Partition) ID() string {
	return p.id
}

// Name returns the name of the partition.
func (p Partition) Name() string {
	return p.name
}

// DNSSuffix returns the base domain name of the partition.
func (p Partition) DNSSuffix() string {
	return p.dnsSuffix
}

// RegionRegex return the regular expression that matches Region IDs for the partition.
func (p Partition) RegionRegex() *regexp.Regexp {
	return p.regionRegex
}

// Regions returns a map of Regions for the partition, indexed by their ID.
func (p Partition) Regions() map[string]Region {
	return maps.Clone(p.regions)
}

// Services returns a map of service endpoints for the partition, indexed by their ID.
func (p Partition) Services() map[string]Service {
	return maps.Clone(p.services)
}

// DefaultPartitions returns a list of the partitions.
func DefaultPartitions() []Partition {
	ps := make([]Partition, 0, len(partitions))

	for _, p := range partitions {
		ps = append(ps, p)
	}

	return ps
}

// PartitionForRegion returns the first partition which includes the specific Region.
func PartitionForRegion(ps []Partition, regionID string) (Partition, bool) {
	for _, p := range ps {
		if _, ok := p.regions[regionID]; ok || p.regionRegex.MatchString(regionID) {
			return p, true
		}
	}

	return Partition{}, false
}

/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	types "github.com/vmware/govcloudair/types/v56"
)

type OrgVDCNetwork struct {
	OrgVDCNetwork *types.OrgVDCNetwork
	c             *Client
}

func NewOrgVDCNetwork(c *Client) *OrgVDCNetwork {
	return &OrgVDCNetwork{
		OrgVDCNetwork: new(types.OrgVDCNetwork),
		c:             c,
	}
}

package azureutil

import (
	"fmt"
)

const (
	fmtNIC = "%s-nic"
	fmtIP  = "%s-ip"
	fmtNSG = "%s-firewall"
	fmtVM  = "%s"
)

// ResourceNaming provides methods to construct Azure resource names for a given
// machine name.
type ResourceNaming string

func (r ResourceNaming) IP() string  { return fmt.Sprintf(fmtIP, r) }
func (r ResourceNaming) NIC() string { return fmt.Sprintf(fmtNIC, r) }
func (r ResourceNaming) NSG() string { return fmt.Sprintf(fmtNSG, r) }
func (r ResourceNaming) VM() string  { return fmt.Sprintf(fmtVM, r) }

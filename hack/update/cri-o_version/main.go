package main

import (
	"strings"
)

func main() {
	crioData, crioErr := detectCrio()
	crictlData, crictlErr := detectCrictl()
	if crictlErr != nil || crioErr != nil {
		return
	}
	// now crio requires crictl which has the same major&minor version with crio
	// so we only update them together when this condition is satisfied
	if strings.HasPrefix(crictlData.Version, "v"+crioData.MMVersion) {
		crioData.updateCrio()
		crictlData.updateCrictl()
	}

}

package driver

import (
	"github.com/blang/semver"
	"github.com/golang/glog"
)

// minHyperkitVersion is the minimum version of the minikube hyperkit driver compatible with the current minikube code
var minHyperkitVersion semver.Version

const minHyperkitVersionStr = "1.10.0"

func init() {
	v, err := semver.New(minHyperkitVersionStr)
	if err != nil {
		glog.Fatalf("Failed to parse the hyperkit driver version: %v", err)
	}
	minHyperkitVersion = *v
}

func minDriverVersion(driver string, mkVer semver.Version) semver.Version {
	switch driver {
	case HyperKit:
		return minHyperkitVersion
	case KVM2:
		return mkVer
	default:
		glog.Warningf("Unexpected driver: %v", driver)
		return mkVer
	}
}

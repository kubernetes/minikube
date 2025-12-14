// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func awsSdkGoUserAgent() string {
	// See https://github.com/aws/aws-sdk-go-v2/blob/4051ca807a0308bc9f169ca308262b328c2692a3/aws/middleware/user_agent_test.go#L18C1-L18C1
	return fmt.Sprintf("%s/%s os/%s lang/go#%s md/GOOS#%s md/GOARCH#%s", aws.SDKName, aws.SDKVersion, getNormalizedOSName(), strings.TrimPrefix(runtime.Version(), "go"), runtime.GOOS, runtime.GOARCH)
}

// Copied from https://github.com/aws/aws-sdk-go-v2/blob/main/aws/middleware/osname.go
func getNormalizedOSName() (os string) {
	switch runtime.GOOS {
	case "android":
		os = "android"
	case "linux":
		os = "linux"
	case "windows":
		os = "windows"
	case "darwin":
		os = "macos"
	case "ios":
		os = "ios"
	default:
		os = "other"
	}
	return os
}

// cleanUserAgent removes:
// * the "api/<whatever>" product that the AWS SDK adds to the user-agent string
// * the "ua/<whatever>" product that contains the User-Agent string version
// * the "m/<whatever>" product that contains the feature flags
func cleanUserAgent(ua string) string {
	var parts []string
	for _, v := range strings.Split(ua, " ") {
		if strings.HasPrefix(v, "api/") {
			continue
		}
		if strings.HasPrefix(v, "ua/") {
			continue
		}
		if strings.HasPrefix(v, "m/") {
			continue
		}
		parts = append(parts, v)
	}
	return strings.Join(parts, " ")
}
